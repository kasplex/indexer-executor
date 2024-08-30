
////////////////////////////////
package storage

import (
    "sync"
    "time"
    "math"
    //"log/slog"
    "github.com/tecbot/gorocksdb"
)

////////////////////////////////
const nWriteRocks = 100
const nGetRocks = 100
const nBatchMaxRocks = 100

////////////////////////////////
var mtsBatchLastRocks = int64(0)

////////////////////////////////
func doGetBatchRocks(lenBatch int, nStart int, fGet func(int, int, *gorocksdb.TransactionDB, *gorocksdb.ReadOptions) (error)) (int64, error) {
    if lenBatch <= 0 {
        return 0, nil
    }
    mtss := time.Now().UnixMilli()
    if nStart == 0 {
        mtsBatchLastRocks = mtss
    }
    nBatch := int(math.Ceil(float64(lenBatch-nStart) / float64(nGetRocks)))
    nStartNext := 0
    if nBatch > nBatchMaxRocks {
        nBatch = nBatchMaxRocks
        nStartNext = nGetRocks * nBatch + nStart
    }
    wg := &sync.WaitGroup{}
    errList := make(chan error, nBatch)
    for i := 0; i < nBatch; i ++ {
        iStart := nStart + i*nGetRocks
        iEnd := nStart + (i+1)*nGetRocks
        if iEnd >= lenBatch {
            iEnd = lenBatch
        }
        wg.Add(1)
        go func() {
            err := fGet(iStart, iEnd, sRuntime.rocksTx, sRuntime.rOptRocks)
            if err != nil {
                errList <- err
            }
            wg.Done()
        }()
    }
    wg.Wait()
    if len(errList) > 0 {
        err := <- errList
        return 0, err
    }
    if nStartNext > 0 {
        _, err := doGetBatchRocks(lenBatch, nStartNext, fGet)
        if err != nil {
            return 0, err
        }
    }
    return time.Now().UnixMilli() - mtss, nil
}
