
////////////////////////////////
package explorer

import (
    "time"
    "strconv"
    "log/slog"
    "kasplex-executor/storage"
    "kasplex-executor/operation"
)

////////////////////////////////
const lenVspcListMax = 200
const lenVspcCheck = 23
const lenRollbackListMax = 30

////////////////////////////////
func scan() {
    mtss := time.Now().UnixMilli()
    
    // Get the next vspc data list.
    vspcLast := storage.DataVspcType{
        DaaScore: eRuntime.cfg.DaaScoreRange[0][0],
    }
    daaScoreStart := vspcLast.DaaScore
    // Use the last vspc if not empty list.
    lenVspcRuntime := len(eRuntime.vspcList)
    if lenVspcRuntime > 0 {
        vspcLast = eRuntime.vspcList[lenVspcRuntime-1]
        daaScoreStart = vspcLast.DaaScore - lenVspcCheck
        if daaScoreStart < eRuntime.vspcList[0].DaaScore {
            daaScoreStart = eRuntime.vspcList[0].DaaScore
        }
    }
    //_, daaScoreStart = checkDaaScoreRange(daaScoreStart)
    // Get next vspc data list from cluster db.
    vspcListNext, mtsBatchVspc, err := storage.GetNodeVspcList(daaScoreStart, lenVspcListMax)
    if err != nil {
        slog.Warn("storage.GetNodeVspcList failed, sleep 3s.", "daaScore", daaScoreStart, "error", err.Error())
        time.Sleep(3000*time.Millisecond)
        return
    }
    // Ignore the last reserved vspc data, reduce the probability of vspc-reorg.
    lenVspcNext := len(vspcListNext)
    lenVspcNext -= eRuntime.cfg.Hysteresis
    if lenVspcNext <= 0 {
        slog.Debug("storage.GetNodeVspcList empty.", "daaScore", daaScoreStart)
        time.Sleep(1550*time.Millisecond)
        return
    }
    vspcListNext = vspcListNext[:lenVspcNext]
    slog.Info("storage.GetNodeVspcList", "daaScore", daaScoreStart, "lenBlock/mSecond", strconv.Itoa(lenVspcNext)+"/"+strconv.Itoa(int(mtsBatchVspc)), "synced", eRuntime.synced)

    // Check vspc list if need rollback.
    daaScoreRollback, vspcListNext := checkRollbackNext(eRuntime.vspcList, vspcListNext, daaScoreStart)
    if daaScoreRollback > 0 {
        daaScoreLast := uint64(0)
        mtsRollback := int64(0)
        // Rollback to the last state data batch.
        lenRollback := len(eRuntime.rollbackList) - 1
        if (lenRollback >= 0 && eRuntime.rollbackList[lenRollback].DaaScoreEnd >= daaScoreRollback) {
            daaScoreLast = eRuntime.rollbackList[lenRollback].DaaScoreStart
            mtsRollback, err = storage.RollbackOpStateBatch(eRuntime.rollbackList[lenRollback])
            if err != nil {
                slog.Warn("storage.RollbackOpStateBatch failed, sleep 3s.", "error", err.Error())
                time.Sleep(3000*time.Millisecond)
                return
            }
            // Remove the vspc data of rollback.
            for {
                lenVspcRuntime = len(eRuntime.vspcList)
                if lenVspcRuntime <= 0 {
                    break
                }
                lenVspcRuntime --
                if eRuntime.vspcList[lenVspcRuntime].DaaScore >= daaScoreLast {
                    if lenVspcRuntime == 0 {
                        eRuntime.vspcList = []storage.DataVspcType{}
                        break
                    }
                    eRuntime.vspcList = eRuntime.vspcList[:lenVspcRuntime]
                    continue
                }
                break
            }
            // Remove the last rollback data.
            eRuntime.rollbackList = eRuntime.rollbackList[:lenRollback]
            storage.SetRuntimeRollbackLast(eRuntime.rollbackList)
        } else {
            eRuntime.vspcList = vspcListNext
        }
        storage.SetRuntimeVspcLast(eRuntime.vspcList)
        slog.Info("explorer.checkRollbackNext", "start/rollback/last", strconv.FormatUint(daaScoreStart,10)+"/"+strconv.FormatUint(daaScoreRollback,10)+"/"+strconv.FormatUint(daaScoreLast,10), "mSecond", strconv.Itoa(int(mtsRollback)))
        return
    } else if vspcListNext == nil {
        slog.Debug("storage.checkDaaScoreRollback empty.", "daaScore", daaScoreStart)
        time.Sleep(1750*time.Millisecond)
        return
    }
    lenVspcNext = len(vspcListNext)
    slog.Debug("explorer.checkRollbackNext", "start/next", strconv.FormatUint(daaScoreStart,10)+"/"+strconv.FormatUint(vspcListNext[0].DaaScore,10))
    
    // Extract and get the transaction list.
    txDataList := []storage.DataTransactionType{}
    for _, vspc := range vspcListNext {
        passed, _ := checkDaaScoreRange(vspc.DaaScore)
        if (!passed || vspc.DaaScore <= vspcLast.DaaScore) {
            continue
        }
        for _, txId := range vspc.TxIdList {
            txDataList = append(txDataList, storage.DataTransactionType{
                TxId: txId,
                DaaScore: vspc.DaaScore,
                BlockAccept: vspc.Hash,
            })
        }
    }
    // Get the transaction data list from cluster db.
    lenTxData := len(txDataList)
    txDataList, mtsBatchTx, err := storage.GetNodeTransactionDataList(txDataList)
    if err != nil {
        slog.Warn("storage.GetNodeTransactionDataList failed, sleep 3s.", "lenTransaction", lenTxData, "error", err.Error())
        time.Sleep(3000*time.Millisecond)
        return
    }
    slog.Info("storage.GetNodeTransactionDataList", "lenTransaction/mSecond", strconv.Itoa(lenTxData)+"/"+strconv.Itoa(int(mtsBatchTx)))
    
    // Parse the transaction and calculate fee for OP.
    opDataList, mtsBatchOp, err := ParseOpDataList(txDataList)
    if err != nil {
        slog.Warn("explorer.ParseOpDataList failed, sleep 3s.", "error", err.Error())
        time.Sleep(3000*time.Millisecond)
        return
    }
    lenOpData := len(opDataList)
    slog.Info("explorer.ParseOpDataList", "lenOperation/mSecond", strconv.Itoa(lenOpData)+"/"+strconv.Itoa(int(mtsBatchOp)))
    
    // Prepare the op data list.
    stateMap, mtsBatchSt, err := operation.PrepareStateBatch(opDataList)
    if err != nil {
        slog.Warn("operation.PrepareStateBatch failed, sleep 3s.", "error", err.Error())
        time.Sleep(3000*time.Millisecond)
        return
    }
    slog.Debug("operation.PrepareStateBatch", "lenToken/lenBalance/mSecond", strconv.Itoa(len(stateMap.StateTokenMap))+"/"+strconv.Itoa(len(stateMap.StateBalanceMap))+"/"+strconv.Itoa(int(mtsBatchSt)))
    
    // Execute the op list and generate the rollback data.
    checkpointLast := ""
    if len(eRuntime.rollbackList) > 0 {
        checkpointLast = eRuntime.rollbackList[len(eRuntime.rollbackList)-1].CheckpointAfter
    }
    rollback, mtsBatchExe, err := operation.ExecuteBatch(opDataList, stateMap, checkpointLast, eRuntime.testnet)
    if err != nil {
        slog.Warn("operation.ExecuteBatch failed, sleep 3s.", "error", err.Error())
        time.Sleep(3000*time.Millisecond)
        return
    }
    rollback.DaaScoreStart = vspcListNext[0].DaaScore
    rollback.DaaScoreEnd = vspcListNext[lenVspcNext-1].DaaScore
    if rollback.CheckpointAfter == "" {
        rollback.CheckpointAfter = rollback.CheckpointBefore
    }
    if rollback.OpScoreLast == 0 {
        rollback.OpScoreLast = eRuntime.opScoreLast
    } else {
        eRuntime.opScoreLast = rollback.OpScoreLast
    }
    slog.Debug("operation.ExecuteBatch", "checkpoint", rollback.CheckpointAfter, "lenOperation/mSecond", strconv.Itoa(lenOpData)+"/"+strconv.Itoa(int(mtsBatchExe)))
    
    // Save the op/state result data list.
    mtsBatchList, err := storage.SaveOpStateBatch(opDataList, stateMap)
    if err != nil {
        slog.Warn("storage.SaveOpStateBatch failed, sleep 3s.", "error", err.Error())
        time.Sleep(3000*time.Millisecond)
        return
    }
    slog.Debug("operation.SaveOpStateBatch", "mSecondList", strconv.Itoa(int(mtsBatchList[0]))+"/"+strconv.Itoa(int(mtsBatchList[1]))+"/"+strconv.Itoa(int(mtsBatchList[2]))+"/"+strconv.Itoa(int(mtsBatchList[3])))
    
    // Update the runtime data.
    eRuntime.synced = false
    if (lenVspcNext < 50) {
        eRuntime.synced = true
    }
    storage.SetRuntimeSynced(eRuntime.synced, eRuntime.opScoreLast, vspcListNext[lenVspcNext-1].DaaScore)
    eRuntime.vspcList = append(eRuntime.vspcList, vspcListNext...)
    lenStart := len(eRuntime.vspcList) - lenVspcListMax
    if lenStart > 0 {
        eRuntime.vspcList = eRuntime.vspcList[lenStart:]
    }
    storage.SetRuntimeVspcLast(eRuntime.vspcList)
    eRuntime.rollbackList = append(eRuntime.rollbackList, rollback)
    lenStart = len(eRuntime.rollbackList) - lenRollbackListMax
    if lenStart > 0 {
        eRuntime.rollbackList = eRuntime.rollbackList[lenStart:]
    }
    storage.SetRuntimeRollbackLast(eRuntime.rollbackList)
        
    // Additional delay if state synced.
    mtsLoop := time.Now().UnixMilli() - mtss
    slog.Info("explorer.scan", "lenRuntimeVspc", len(eRuntime.vspcList), "lenRuntimeRollback", len(eRuntime.rollbackList), "lenOperation", lenOpData, "mSecondLoop", mtsLoop)
    if (eRuntime.synced) {
        mtsLoop = 1650 - mtsLoop
        if mtsLoop <=0 {
            return
        }
        time.Sleep(time.Duration(mtsLoop)*time.Millisecond)
    }
}

////////////////////////////////
func checkRollbackNext(vspcListPrev []storage.DataVspcType, vspcListNext []storage.DataVspcType, daaScoreStart uint64) (uint64, []storage.DataVspcType) {
    if len(vspcListPrev) <= 0 {
        return 0, vspcListNext
    }
    vspcList1 := []storage.DataVspcType{}
    vspcList2 := []storage.DataVspcType{}
    for _, vspc := range vspcListPrev {
        if vspc.DaaScore < daaScoreStart {
            continue
        }   
        vspcList1 = append(vspcList1, vspc)
    }
    lenCheck := len(vspcList1)
    if lenCheck > 0 {
        if len(vspcListNext) <= lenCheck {
            return 0, nil
        } else {
            vspcList2 = vspcListNext[:lenCheck]
        }
    } else {
        return 0, vspcListNext
    }
    for i := 0; i < lenCheck; i ++ {
        if (vspcList1[i].DaaScore != vspcList2[i].DaaScore || vspcList1[i].Hash != vspcList2[i].Hash) {
            return vspcList1[i].DaaScore, vspcListPrev[:(len(vspcListPrev)-lenCheck+i)]
        }
    }
    return 0, vspcListNext[lenCheck:]
}

////////////////////////////////
func checkDaaScoreRange(daaScore uint64) (bool, uint64) {
    for _, dRange := range eRuntime.cfg.DaaScoreRange {
        if daaScore < dRange[0] {
            return false, dRange[0]
        } else if (daaScore <= dRange[1]) {
            return true, daaScore
        }
    }
    return false, daaScore
}
