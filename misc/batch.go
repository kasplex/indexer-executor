
////////////////////////////////
package misc

import (
    "time"
    "sync"
    "math"
)

////////////////////////////////
const nGoroutine = 100

////////////////////////////////
func GoBatch(lenBatch int, fGo func(int) (error)) (int64, error) {
    if lenBatch <= 0 {
        return 0, nil
    }
    mtss := time.Now().UnixMilli()
    nBatch := int(math.Ceil(float64(lenBatch) / float64(nGoroutine)))
    wg := &sync.WaitGroup{}
    errList := make(chan error, nBatch)
    for i := 0; i < nGoroutine; i ++ {
        wg.Add(1)
        go func() {
            for j := i*nBatch; j < (i+1)*nBatch; j ++ {
                if j >= lenBatch {
                    break
                }
                err := fGo(j)
                if err != nil {
                    errList <- err
                    break
                }
            }
            wg.Done()
        }()
    }
    wg.Wait()
    if len(errList) > 0 {
        err := <- errList
        return 0, err
    }
    return time.Now().UnixMilli() - mtss, nil
}
