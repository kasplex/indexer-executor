
////////////////////////////////
package storage

import (
    "sync"
    "time"
    "strings"
    //"log/slog"
    "encoding/json"
    "github.com/gocql/gocql"
    "github.com/tecbot/gorocksdb"
)

////////////////////////////////
const OpRangeBy = uint64(100000)

////////////////////////////////
const KeyPrefixStateToken = "sttoken_"
const KeyPrefixStateBalance = "stbalance_"
const KeyPrefixStateMarket = "stmarket_"
const KeyPrefixStateBlacklist = "stblacklist_"
// KeyPrefixStateXxx ...

////////////////////////////////
func GetStateTokenMap(tokenMap map[string]*StateTokenType) (int64, error) {
    keyList := [][]byte{}
    for tick := range tokenMap {
        keyList = append(keyList, []byte(KeyPrefixStateToken+tick))
    }
    mutex := new(sync.RWMutex)
    mtsBatch, err := doGetBatchRocks(len(keyList), 0, func(iStart int, iEnd int, rdb *gorocksdb.TransactionDB, rro *gorocksdb.ReadOptions) (error) {
        for i := iStart; i < iEnd; i ++ {
            row, err := rdb.Get(rro, keyList[i])
            if err != nil {
                return err
            }
            dataByte := row.Data()
            if dataByte == nil {
                continue
            }
            decoded := StateTokenType{}
            err = json.Unmarshal(dataByte, &decoded)
            if err != nil {
                return err
            }
            mutex.Lock()
            tokenMap[decoded.Tick] = &decoded
            mutex.Unlock()
        }
        return nil
    })
    if err != nil {
        return 0, err
    }
    return mtsBatch, nil
}

////////////////////////////////
func GetStateBalanceMap(balanceMap map[string]*StateBalanceType) (int64, error) {
    keyList := [][]byte{}
    for addrTick := range balanceMap {
        keyList = append(keyList, []byte(KeyPrefixStateBalance+addrTick))
    }
    mutex := new(sync.RWMutex)
    mtsBatch, err := doGetBatchRocks(len(keyList), 0, func(iStart int, iEnd int, rdb *gorocksdb.TransactionDB, rro *gorocksdb.ReadOptions) (error) {
        for i := iStart; i < iEnd; i ++ {
            row, err := rdb.Get(rro, keyList[i])
            if err != nil {
                return err
            }
            dataByte := row.Data()
            if dataByte == nil {
                continue
            }
            decoded := StateBalanceType{}
            err = json.Unmarshal(dataByte, &decoded)
            if err != nil {
                return err
            }
            mutex.Lock()
            balanceMap[decoded.Address+"_"+decoded.Tick] = &decoded
            mutex.Unlock()
        }
        return nil
    })
    if err != nil {
        return 0, err
    }
    return mtsBatch, nil
}

////////////////////////////////
func GetStateMarketMap(marketMap map[string]*StateMarketType) (int64, error) {
    keyList := [][]byte{}
    for tickAddrTxid := range marketMap {
        keyList = append(keyList, []byte(KeyPrefixStateMarket+tickAddrTxid))
    }
    mutex := new(sync.RWMutex)
    mtsBatch, err := doGetBatchRocks(len(keyList), 0, func(iStart int, iEnd int, rdb *gorocksdb.TransactionDB, rro *gorocksdb.ReadOptions) (error) {
        for i := iStart; i < iEnd; i ++ {
            row, err := rdb.Get(rro, keyList[i])
            if err != nil {
                return err
            }
            dataByte := row.Data()
            if dataByte == nil {
                continue
            }
            decoded := StateMarketType{}
            err = json.Unmarshal(dataByte, &decoded)
            if err != nil {
                return err
            }
            mutex.Lock()
            marketMap[decoded.Tick+"_"+decoded.TAddr+"_"+decoded.UTxId] = &decoded
            mutex.Unlock()
        }
        return nil
    })
    if err != nil {
        return 0, err
    }
    return mtsBatch, nil
}

////////////////////////////////
func GetStateBlacklistMap(blacklistMap map[string]*StateBlacklistType) (int64, error) {
    keyList := [][]byte{}
    for tickAddr := range blacklistMap {
        keyList = append(keyList, []byte(KeyPrefixStateBlacklist+tickAddr))
    }
    mutex := new(sync.RWMutex)
    mtsBatch, err := doGetBatchRocks(len(keyList), 0, func(iStart int, iEnd int, rdb *gorocksdb.TransactionDB, rro *gorocksdb.ReadOptions) (error) {
        for i := iStart; i < iEnd; i ++ {
            row, err := rdb.Get(rro, keyList[i])
            if err != nil {
                return err
            }
            dataByte := row.Data()
            if dataByte == nil {
                continue
            }
            decoded := StateBlacklistType{}
            err = json.Unmarshal(dataByte, &decoded)
            if err != nil {
                return err
            }
            mutex.Lock()
            blacklistMap[decoded.Tick+"_"+decoded.Address] = &decoded
            mutex.Unlock()
        }
        return nil
    })
    if err != nil {
        return 0, err
    }
    return mtsBatch, nil
}

////////////////////////////////
// GetStateXxx ...

////////////////////////////////
func CopyDataStateMap(stateMapFrom DataStateMapType, stateMapTo *DataStateMapType) {
    stateMapTo.StateTokenMap = make(map[string]*StateTokenType)
    stateMapTo.StateBalanceMap = make(map[string]*StateBalanceType)
    stateMapTo.StateMarketMap = make(map[string]*StateMarketType)
    stateMapTo.StateBlacklistMap = make(map[string]*StateBlacklistType)
    // stateMapTo.StateXxxMap ...
    for key, stToken := range stateMapFrom.StateTokenMap {
        if stToken == nil {
            stateMapTo.StateTokenMap[key] = nil
            continue
        }
        stData := *stToken
        stateMapTo.StateTokenMap[key] = &stData
    }
    for key, stBalance := range stateMapFrom.StateBalanceMap {
        if stBalance == nil {
            stateMapTo.StateBalanceMap[key] = nil
            continue
        }
        stData := *stBalance
        stateMapTo.StateBalanceMap[key] = &stData
    }
    for key, stMarket := range stateMapFrom.StateMarketMap {
        if stMarket == nil {
            stateMapTo.StateMarketMap[key] = nil
            continue
        }
        stData := *stMarket
        stateMapTo.StateMarketMap[key] = &stData
    }
    for key, stBlacklist := range stateMapFrom.StateBlacklistMap {
        if stBlacklist == nil {
            stateMapTo.StateBlacklistMap[key] = nil
            continue
        }
        stData := *stBlacklist
        stateMapTo.StateBlacklistMap[key] = &stData
    }
    // StateXxx ...
}

////////////////////////////////
func SaveStateBatchCassa(stateMap DataStateMapType) (int64, error) {
    mtss := time.Now().UnixMilli()
    keyList := make([]string, 0, len(stateMap.StateTokenMap))
    for key := range stateMap.StateTokenMap {
        keyList = append(keyList, key)
    }
    _, err := startExecuteBatchCassa(len(keyList), func(batch *gocql.Batch, i int) (error) {
        stToken := stateMap.StateTokenMap[keyList[i]]
        tick := keyList[i]
        if stToken == nil {
            batch.Query(cqlnDeleteStateToken, tick[:2], tick)
            return nil
        }
        meta := &StateTokenMetaType{
            Max: stToken.Max,
            Lim: stToken.Lim,
            Pre: stToken.Pre,
            Dec: stToken.Dec,
            From: stToken.From,
            To: stToken.To,
            Name: stToken.Name,
            TxId: stToken.TxId,
            OpAdd: stToken.OpAdd,
            MtsAdd: stToken.MtsAdd,
        }
        metaJson, _ := json.Marshal(meta)
        batch.Query(cqlnSaveStateToken, tick[:2], tick, string(metaJson), stToken.Minted, stToken.OpMod, stToken.MtsMod, stToken.Mod, stToken.Burned)
        return nil
    })
    if err != nil {
        return 0, err
    }
    keyList = make([]string, 0, len(stateMap.StateBalanceMap))
    for key := range stateMap.StateBalanceMap {
        keyList = append(keyList, key)
    }
    _, err = startExecuteBatchCassa(len(keyList), func(batch *gocql.Batch, i int) (error) {
        stBalance := stateMap.StateBalanceMap[keyList[i]]
        key := strings.Split(keyList[i], "_")
        if stBalance == nil {
            batch.Query(cqlnDeleteStateBalance, key[0], key[1])
            return nil
        }
        batch.Query(cqlnSaveStateBalance, key[0], key[1], stBalance.Dec, stBalance.Balance, stBalance.Locked, stBalance.OpMod)
        return nil
    })
    if err != nil {
        return 0, err
    }
    keyList = make([]string, 0, len(stateMap.StateMarketMap))
    for key := range stateMap.StateMarketMap {
        keyList = append(keyList, key)
    }
    _, err = startExecuteBatchCassa(len(keyList), func(batch *gocql.Batch, i int) (error) {
        stMarket := stateMap.StateMarketMap[keyList[i]]
        key := strings.Split(keyList[i], "_")
        if stMarket == nil {
            batch.Query(cqlnDeleteStateMarket, key[0], key[1]+"_"+key[2])
            return nil
        }
        batch.Query(cqlnSaveStateMarket, key[0], key[1]+"_"+key[2], stMarket.UAddr, stMarket.UAmt, stMarket.UScript, stMarket.TAmt, stMarket.OpAdd)
        return nil
    })
    if err != nil {
        return 0, err
    }
    keyList = make([]string, 0, len(stateMap.StateBlacklistMap))
    for key := range stateMap.StateBlacklistMap {
        keyList = append(keyList, key)
    }
    _, err = startExecuteBatchCassa(len(keyList), func(batch *gocql.Batch, i int) (error) {
        stBlacklist := stateMap.StateBlacklistMap[keyList[i]]
        key := strings.Split(keyList[i], "_")
        if stBlacklist == nil {
            batch.Query(cqlnDeleteStateBlacklist, key[0], key[1])
            return nil
        }
        batch.Query(cqlnSaveStateBlacklist, key[0], key[1], stBlacklist.OpAdd)
        return nil
    })
    if err != nil {
        return 0, err
    }
    // StateXxx ...
    return time.Now().UnixMilli() - mtss, nil
}

////////////////////////////////
func SaveOpDataBatchCassa(opDataList []DataOperationType) (int64, error) {
    mtss := time.Now().UnixMilli()
    stateJsonMap := make(map[string]string, len(opDataList))
    scriptJsonMap := make(map[string]string, len(opDataList))
    _, err := startExecuteBatchCassa(len(opDataList), func(batch *gocql.Batch, i int) (error) {
        state := &DataOpStateType{
            BlockAccept: opDataList[i].BlockAccept,
            Fee: opDataList[i].Fee,
            FeeLeast: opDataList[i].FeeLeast,
            MtsAdd: opDataList[i].MtsAdd,
            OpScore: opDataList[i].OpScore,
            OpAccept: opDataList[i].OpAccept,
            OpError: opDataList[i].OpError,
            Checkpoint: opDataList[i].Checkpoint,
        }
        stateJson, _ := json.Marshal(state)
        scriptJson, _ := json.Marshal(opDataList[i].OpScript[0])
        stBeforeJson, _ := json.Marshal(opDataList[i].StBefore)
        stAfterJson, _ := json.Marshal(opDataList[i].StAfter)
        stateJsonMap[opDataList[i].TxId] = string(stateJson)
        scriptJsonMap[opDataList[i].TxId] = string(scriptJson)
        batch.Query(cqlnSaveOpData, opDataList[i].TxId, string(stateJson), string(scriptJson), string(stBeforeJson), string(stAfterJson))
        return nil
    })
    if err != nil {
        return 0, err
    }
    _, err = startExecuteBatchCassa(len(opDataList), func(batch *gocql.Batch, i int) (error) {
        tickAffc := strings.Join(opDataList[i].SsInfo.TickAffc, ",")
        addressAffc := strings.Join(opDataList[i].SsInfo.AddressAffc, ",")
        // xxxAffc ...
        opRange := opDataList[i].OpScore / OpRangeBy
        batch.Query(cqlnSaveOpList, opRange, opDataList[i].OpScore, opDataList[i].TxId, stateJsonMap[opDataList[i].TxId], scriptJsonMap[opDataList[i].TxId], tickAffc, addressAffc)
        return nil
    })
    if err != nil {
        return 0, err
    }
    return time.Now().UnixMilli() - mtss, nil
}

////////////////////////////////
func DeleteOpDataBatchCassa(opScoreList []uint64, txIdList []string) (int64, error) {
    mtss := time.Now().UnixMilli()
    _, err := startExecuteBatchCassa(len(opScoreList), func(batch *gocql.Batch, i int) (error) {
        opRange := opScoreList[i] / OpRangeBy
        batch.Query(cqlnDeleteOpList, opRange, opScoreList[i])
        return nil
    })
    if err != nil {
        return 0, err
    }
    _, err = startExecuteBatchCassa(len(txIdList), func(batch *gocql.Batch, i int) (error) {
        batch.Query(cqlnDeleteOpData, txIdList[i])
        return nil
    })
    if err != nil {
        return 0, err
    }
    return time.Now().UnixMilli() - mtss, nil
}

////////////////////////////////
func SaveStateBatchRocksBegin(stateMap DataStateMapType, txRocks *gorocksdb.Transaction) (*gorocksdb.Transaction, int64, error) {
    mtss := time.Now().UnixMilli()
    if txRocks == nil {
        txRocks = sRuntime.rocksTx.TransactionBegin(sRuntime.wOptRocks, sRuntime.txOptRocks, nil)
    }
    var err error
    var valueJson []byte
    for key, token := range stateMap.StateTokenMap {
        key = KeyPrefixStateToken + key
        if token == nil {
            err = txRocks.Delete([]byte(key))
        } else {
            valueJson, _ = json.Marshal(token)
            err = txRocks.Put([]byte(key), valueJson)
        }
        if err != nil {
            txRocks.Rollback()
            return txRocks, 0, err
        }
    }
    for key, balance := range stateMap.StateBalanceMap {
        key = KeyPrefixStateBalance + key
        if balance == nil {
            err = txRocks.Delete([]byte(key))
        } else {
            valueJson, _ = json.Marshal(balance)
            err = txRocks.Put([]byte(key), valueJson)
        }
        if err != nil {
            txRocks.Rollback()
            return txRocks, 0, err
        }
    }
    for key, market := range stateMap.StateMarketMap {
        key = KeyPrefixStateMarket + key
        if market == nil {
            err = txRocks.Delete([]byte(key))
        } else {
            valueJson, _ = json.Marshal(market)
            err = txRocks.Put([]byte(key), valueJson)
        }
        if err != nil {
            txRocks.Rollback()
            return txRocks, 0, err
        }
    }
    for key, blacklist := range stateMap.StateBlacklistMap {
        key = KeyPrefixStateBlacklist + key
        if blacklist == nil {
            err = txRocks.Delete([]byte(key))
        } else {
            valueJson, _ = json.Marshal(blacklist)
            err = txRocks.Put([]byte(key), valueJson)
        }
        if err != nil {
            txRocks.Rollback()
            return txRocks, 0, err
        }
    }
    // StateXxx ...
    return txRocks, time.Now().UnixMilli() - mtss, nil
}

////////////////////////////////
func SaveOpStateBatch(opDataList []DataOperationType, stateMap DataStateMapType) ([]int64, error) {
    mtsBatchList := [4]int64{}
    mtsBatchList[0] = time.Now().UnixMilli()
    txRocks, _, err := SaveStateBatchRocksBegin(stateMap, nil)
    defer txRocks.Destroy()
    if err != nil {
        return nil, err
    }
    mtsBatchList[1] = time.Now().UnixMilli()
    _, err = SaveStateBatchCassa(stateMap)
    if err != nil {
        txRocks.Rollback()
        return nil, err
    }
    mtsBatchList[2] = time.Now().UnixMilli()
    _, err = SaveOpDataBatchCassa(opDataList)
    if err != nil {
        txRocks.Rollback()
        return nil, err
    }
    mtsBatchList[3] = time.Now().UnixMilli()
    err = txRocks.Commit()
    if err != nil {
        txRocks.Rollback()
        return nil, err
    }
    mtsBatchList[0] = mtsBatchList[1] - mtsBatchList[0]
    mtsBatchList[1] = mtsBatchList[2] - mtsBatchList[1]
    mtsBatchList[2] = mtsBatchList[3] - mtsBatchList[2]
    mtsBatchList[3] = time.Now().UnixMilli() - mtsBatchList[3]
    return mtsBatchList[:], nil
}

////////////////////////////////
func RollbackOpStateBatch(rollback DataRollbackType) (int64, error) {
    mtss := time.Now().UnixMilli()
    txRocks, _, err := SaveStateBatchRocksBegin(rollback.StateMapBefore, nil)
    defer txRocks.Destroy()
    if err != nil {
        return 0, err
    }
    _, err = SaveStateBatchCassa(rollback.StateMapBefore)
    if err != nil {
        txRocks.Rollback()
        return 0, err
    }
    _, err = DeleteOpDataBatchCassa(rollback.OpScoreList, rollback.TxIdList)
    if err != nil {
        txRocks.Rollback()
        return 0, err
    }
    err = txRocks.Commit()
    if err != nil {
        txRocks.Rollback()
        return 0, err
    }
    return time.Now().UnixMilli() - mtss, nil
}
