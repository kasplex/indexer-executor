
////////////////////////////////
package storage

import (
    "sync"
    "time"
    "math"
    "log/slog"
    "github.com/gocql/gocql"
)

////////////////////////////////
const numConns = 10
const inMaxCassa = 10
const nQueryCassa = 5
const nBatchMaxCassa = 100
const mtsDelayExecuteCassa = 10
const mtsDelayQueryCassa = 5

////////////////////////////////
var mtsBatchLastCassa = int64(0)

////////////////////////////////
func startExecuteBatchCassa(lenBatch int, fAdd func(*gocql.Batch, int) (error)) (int64, error) {
    if lenBatch <= 0 {
        return 0, nil
    }
    mtss := time.Now().UnixMilli()
    mtsBatchLastCassa = mtss
    nStart := 0
    nBatchAdj := nBatchMaxCassa
    mtsDelay := mtsDelayExecuteCassa
    nRetry := 0
    for {
        nStartNext, err := doExecuteBatchCassa(lenBatch, nStart, fAdd, nBatchAdj)
        if err != nil {
            nRetry ++
            if nRetry > nBatchMaxCassa {
                return 0, err
            }
            nBatchAdj --
            if nBatchAdj < 1 {
                nBatchAdj = 1
            }
            mtsDelay = mtsDelay * (10+nRetry*2) / 10
            if mtsDelay > 1000 {
                mtsDelay = 1000
            }
        } else {
            nStart = nStartNext
            nRetry = 0
            nBatchAdj += 3
            if nBatchAdj > nBatchMaxCassa {
                nBatchAdj = nBatchMaxCassa
            }
            mtsDelay = mtsDelay * 8 / 10
            if mtsDelay < mtsDelayExecuteCassa {
                mtsDelay = mtsDelayExecuteCassa
            }
        }
        if nStart < 0 {
            break
        }
        mtsNow := time.Now().UnixMilli()
        if mtsNow - mtsBatchLastCassa >= 1000 {
            mtsBatchLastCassa = mtsNow
            slog.Debug("storage.doExecuteBatchCassa", "nStartNext", nStart)
        }
        time.Sleep(time.Duration(mtsDelay) * time.Millisecond)
    }
    return time.Now().UnixMilli() - mtss, nil
}

////////////////////////////////
func doExecuteBatchCassa(lenBatch int, nStart int, fAdd func(*gocql.Batch, int) (error), nBatchAdj int) (int, error) {
    nBatch := int(math.Ceil(float64(lenBatch-nStart) / float64(nQueryCassa)))
    nStartNext := -1
    if nBatch > nBatchAdj {
        nBatch = nBatchAdj
        nStartNext = nQueryCassa * nBatch + nStart
    }
    wg := &sync.WaitGroup{}
    errList := make(chan error, nBatch)
    for i := 0; i < nBatch; i ++ {
        batch := sRuntime.sessionCassa.NewBatch(gocql.UnloggedBatch)
        for j := nStart+i*nQueryCassa; j < nStart+(i+1)*nQueryCassa; j ++ {
            if j >= lenBatch {
                break
            }
            err := fAdd(batch, j)
            if err != nil {
                return nStartNext, err
            }
        }
        wg.Add(1)
        go func() {
            err := sRuntime.sessionCassa.ExecuteBatch(batch)
            if err != nil {
                errList <- err
            }
            wg.Done()
        }()
    }
    wg.Wait()
    if len(errList) > 0 {
        err := <- errList
        return nStartNext, err
    }
    return nStartNext, nil
}

////////////////////////////////
func startQueryBatchInCassa(lenBatch int, fQuery func(int, int, *gocql.Session) (error)) (int64, error) {
    if lenBatch <= 0 {
        return 0, nil
    }
    mtss := time.Now().UnixMilli()
    mtsBatchLastCassa = mtss
    nStart := 0
    nBatchAdj := nBatchMaxCassa
    mtsDelay := mtsDelayQueryCassa
    nRetry := 0
    for {
        nStartNext, err := doQueryBatchInCassa(lenBatch, nStart, fQuery, nBatchAdj)
        if err != nil {
            nRetry ++
            if nRetry > nBatchMaxCassa {
                return 0, err
            }
            nBatchAdj --
            if nBatchAdj < 1 {
                nBatchAdj = 1
            }
            mtsDelay = mtsDelay * (10+nRetry*2) / 10
            if mtsDelay > 1000 {
                mtsDelay = 1000
            }
        } else {
            nStart = nStartNext
            nRetry = 0
            nBatchAdj += 3
            if nBatchAdj > nBatchMaxCassa {
                nBatchAdj = nBatchMaxCassa
            }
            mtsDelay = mtsDelay * 8 / 10
            if mtsDelay < mtsDelayQueryCassa {
                mtsDelay = mtsDelayQueryCassa
            }
        }
        if nStart < 0 {
            break
        }
        mtsNow := time.Now().UnixMilli()
        if mtsNow - mtsBatchLastCassa >= 1000 {
            mtsBatchLastCassa = mtsNow
            slog.Debug("storage.doQueryBatchInCassa", "nStartNext", nStart)
        }
        time.Sleep(time.Duration(mtsDelay) * time.Millisecond)
    }
    return time.Now().UnixMilli() - mtss, nil
}

////////////////////////////////
func doQueryBatchInCassa(lenBatch int, nStart int, fQuery func(int, int, *gocql.Session) (error), nBatchAdj int) (int, error) {
    nBatch := int(math.Ceil(float64(lenBatch-nStart) / float64(inMaxCassa)))
    nStartNext := -1
    if nBatch > nBatchMaxCassa {
        nBatch = nBatchMaxCassa
        nStartNext = inMaxCassa * nBatch + nStart
    }
    wg := &sync.WaitGroup{}
    errList := make(chan error, nBatch)
    for i := 0; i < nBatch; i ++ {
        iStart := nStart + i*inMaxCassa
        iEnd := nStart + (i+1)*inMaxCassa
        if iEnd >= lenBatch {
            iEnd = lenBatch
        }
        wg.Add(1)
        go func() {
            err := fQuery(iStart, iEnd, sRuntime.sessionCassa)
            if err != nil {
                errList <- err
            }
            wg.Done()
        }()
    }
    wg.Wait()
    if len(errList) > 0 {
        err := <- errList
        return nStartNext, err
    }
    return nStartNext, nil
}