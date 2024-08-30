
////////////////////////////////
package storage

import (
    "strconv"
    "encoding/json"
)

////////////////////////////////
const keyPrefixRuntime = "RTA_"  // runtime-arguments
const keyPrefixRuntimeCassa = "EXE_"  // runtime-arguments in the cluster db.

////////////////////////////////
// Get runtime data by key, in the local db.
func GetRuntimeRocks(key string) ([]byte, error) {
    key = keyPrefixRuntime + key
    row, err := sRuntime.rocksTx.Get(sRuntime.rOptRocks, []byte(key))
    if err != nil {
        return nil, err
    }
    return row.Data(), nil
}

////////////////////////////////
// Set runtime data by key, in the local db.
func SetRuntimeRocks(key string, valueJson []byte) (error) {
    key = keyPrefixRuntime + key
    err := sRuntime.rocksTx.Put(sRuntime.wOptRocks, []byte(key), valueJson)
    return err
}

////////////////////////////////
// Get the last processed vspc data list.
func GetRuntimeVspcLast() ([]DataVspcType, error) {
    valueJson, err := GetRuntimeRocks("VSPCLAST")
    if err != nil {
        return nil, err
    }
    if len(valueJson) <= 0 {
        return nil, nil
    }
    list := []DataVspcType{}
    err = json.Unmarshal(valueJson, &list)
    if err != nil {
        return nil, err
    }
    return list, err
}

////////////////////////////////
// Set the last processed vspc data list.
func SetRuntimeVspcLast(list []DataVspcType) (error) {
    valueJson, _ := json.Marshal(list)
    err := SetRuntimeRocks("VSPCLAST", valueJson)
    return err
}

////////////////////////////////
// Get the last rollback data list.
func GetRuntimeRollbackLast() ([]DataRollbackType, error) {
    valueJson, err := GetRuntimeRocks("ROLLBACKLAST")
    if err != nil {
        return nil, err
    }
    if len(valueJson) <= 0 {
        return nil, nil
    }
    list := []DataRollbackType{}
    err = json.Unmarshal(valueJson, &list)
    if err != nil {
        return nil, err
    }
    return list, err
}

////////////////////////////////
// Set the last op data list.
func SetRuntimeRollbackLast(list []DataRollbackType) (error) {
    valueJson, _ := json.Marshal(list)
    err := SetRuntimeRocks("ROLLBACKLAST", valueJson)
    return err
}

////////////////////////////////
// Get runtime data from table "runtime", in the cluster db.
func GetRuntimeCassa(key string) (string, string, string, error) {
    key = keyPrefixRuntimeCassa + key
    row := sRuntime.sessionCassa.Query(cqlnGetRuntime, key)
    defer row.Release()
    var k0, v1, v2, v3 string
    err := row.Scan(&k0, &v1, &v2, &v3)
    if err != nil {
        if err.Error() == "not found"{
            return "", "", "", nil
        }
        return "", "", "", err
    }
    return v1, v2, v3, nil
}

////////////////////////////////
// Set runtime data to table "runtime", in the cluster db.
func SetRuntimeCassa(key string, v1 string, v2 string, v3 string) (error) {
    key = keyPrefixRuntimeCassa + key
    err := sRuntime.sessionCassa.Query(cqlnSetRuntime, key, v1, v2, v3).Exec()
    return err
}

////////////////////////////////
// Get the sync state.
func GetRuntimeSynced() (bool, uint64, error) {
    Synced, _, strDaaScore, err := GetRuntimeCassa("SYNCED")
    if err != nil {
        return false, 0, err
    }
    daaScore, _ := strconv.ParseUint(strDaaScore, 10, 64)
    if Synced == "" {
        return false, daaScore, nil
    }
    return true, daaScore, nil
}

////////////////////////////////
// Set the sync state.
func SetRuntimeSynced(Synced bool, opScore uint64, daaScore uint64) (error) {
    var err error
    strDaaScore := strconv.FormatUint(daaScore, 10)
    strOpScore := strconv.FormatUint(opScore, 10)
    if Synced {
        err = SetRuntimeCassa("SYNCED", "1", strOpScore, strDaaScore)
    } else {
        err = SetRuntimeCassa("SYNCED", "", strOpScore, strDaaScore)
    }
    return err
}

// ...
