
////////////////////////////////
package explorer

import (
    "sync"
    "context"
    "time"
    "log"
    "log/slog"
    "kasplex-executor/config"
    "kasplex-executor/storage"
    "kasplex-executor/operation"
)

////////////////////////////////
type runtimeType struct {
    ctx context.Context
    wg *sync.WaitGroup
    cfg config.StartupConfig
    vspcList []storage.DataVspcType
    rollbackList []storage.DataRollbackType
    opScoreLast uint64
    synced bool
    testnet bool
}
var eRuntime runtimeType

// Available daaScore range.
var daaScoreRange = [][2]uint64{
    {83441551, 83525600},
    {93441551, 18446744073709551615},
}

////////////////////////////////
func Init(ctx context.Context, wg *sync.WaitGroup, cfg config.StartupConfig, testnet bool) {
    slog.Info("explorer.Init start.")
    var err error
    eRuntime.synced = false
    eRuntime.ctx = ctx
    eRuntime.wg = wg
    eRuntime.cfg = cfg
    eRuntime.testnet = testnet
    if eRuntime.cfg.Hysteresis < 0 {
        eRuntime.cfg.Hysteresis = 0
    } else if eRuntime.cfg.Hysteresis > 10 {
        eRuntime.cfg.Hysteresis = 10
    }
    if (!testnet || len(eRuntime.cfg.DaaScoreRange) <= 0) {
        eRuntime.cfg.DaaScoreRange = daaScoreRange
    }
    if len(eRuntime.cfg.TickReserved) > 0 {
        operation.ApplyTickReserved(eRuntime.cfg.TickReserved)
    }
    eRuntime.rollbackList, err = storage.GetRuntimeRollbackLast()
    if err != nil {
        log.Fatalln("explorer.Init fatal:", err.Error())
    }
    eRuntime.vspcList, err = storage.GetRuntimeVspcLast()
    if err != nil {
        log.Fatalln("explorer.Init fatal:", err.Error())
    }
    indexRollback := len(eRuntime.rollbackList) - 1
    if indexRollback >= 0 {
        eRuntime.opScoreLast = eRuntime.rollbackList[indexRollback].OpScoreLast
    }
    if len(eRuntime.vspcList) > 0 {
        lenVspc := len(eRuntime.vspcList)
        vspcLast := eRuntime.vspcList[lenVspc-1]
        slog.Info("explorer.Init", "lastVspcDaaScore", vspcLast.DaaScore, "lastVspcBlockHash", vspcLast.Hash)
        storage.SetRuntimeSynced(false, eRuntime.opScoreLast, vspcLast.DaaScore)
    } else {
        slog.Info("explorer.Init", "lastVspcDaaScore", eRuntime.cfg.DaaScoreRange[0][0], "lastVspcBlockHash", "")
        storage.SetRuntimeSynced(false, eRuntime.opScoreLast, eRuntime.cfg.DaaScoreRange[0][0])
    }
    slog.Info("explorer ready.")
}

////////////////////////////////
func Run() {
    eRuntime.wg.Add(1)
    defer eRuntime.wg.Done()
loop:
    for {
        select {
            case <-eRuntime.ctx.Done():
                slog.Info("explorer.Scan stopped.")
                break loop
            default:
                scan()
                // Basic loop delay.
                time.Sleep(100*time.Millisecond)
        }
    }
}
