
////////////////////////////////
package operation

import (
    "fmt"
    "time"
    "strconv"
    "strings"
    //"log/slog"
    "math/big"
    "golang.org/x/crypto/blake2b"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethod interface {
    Validate(*storage.DataScriptType, bool) (bool)
    FeeLeast(uint64) (uint64)
    PrepareStateKey(*storage.DataOperationType, storage.DataStateMapType)
    Do(*storage.DataOperationType, storage.DataStateMapType, bool) (error)
    UnDo() (error)
    // ...
}

////////////////////////////////
var P_Registered = map[string]bool{}
var Op_Registered = map[string]bool{}
var Method_Registered = map[string]OpMethod{}

////////////////////////////////
var TickIgnored = map[string]bool{
    "KASPA": true, "KASPLX": true,
    "KASP": true, "WKAS": true,
    "GIGA": true, "WBTC": true,
    "WETH": true, "USDT": true,
    "USDC": true, "FDUSD": true,
    "USDD": true, "TUSD": true,
    "USDP": true, "PYUSD": true,
    "EURC": true, "BUSD": true,
    "GUSD": true, "EURT": true,
    "XAUT": true, "TETHER": true,
    
    // ...
}

////////////////////////////////
var TickReserved = map[string]string{}

////////////////////////////////
func ApplyTickReserved(reservedList []string) {
    for _, reserved := range reservedList {
        tickAddr := strings.Split(reserved, "_")
        if len(tickAddr) < 2 {
            continue
        }
        TickReserved[tickAddr[0]] = tickAddr[1]
    }
}

////////////////////////////////
func PrepareStateBatch(opDataList []storage.DataOperationType) (storage.DataStateMapType, int64, error) {
    mtss := time.Now().UnixMilli()
    stateMap := storage.DataStateMapType{
        StateTokenMap: make(map[string]*storage.StateTokenType),
        StateBalanceMap: make(map[string]*storage.StateBalanceType),
        // StateXxx ...
    }
    for _, opData := range opDataList{
        Method_Registered[opData.OpScript.Op].PrepareStateKey(&opData, stateMap)
    }
    _, err := storage.GetStateTokenMap(stateMap.StateTokenMap)
    if err != nil {
        return storage.DataStateMapType{}, 0, err
    }
    _, err = storage.GetStateBalanceMap(stateMap.StateBalanceMap)
    if err != nil {
        return storage.DataStateMapType{}, 0, err
    }
    // GetStateXxx ...
    return stateMap, time.Now().UnixMilli() - mtss, nil
}

////////////////////////////////
func ExecuteBatch(opDataList []storage.DataOperationType, stateMap storage.DataStateMapType, checkpointLast string, testnet bool) (storage.DataRollbackType, int64, error) {
    mtss := time.Now().UnixMilli()
    rollback := storage.DataRollbackType{
        CheckpointBefore: checkpointLast,
        OpScoreList: []uint64{},
        TxIdList: []string{},
    }
    if len(opDataList) <= 0 {
        return rollback, 0, nil
    }
    storage.CopyDataStateMap(stateMap, &rollback.StateMapBefore)
    for i := range opDataList {
        opData := &opDataList[i]
        err := Method_Registered[opData.OpScript.Op].Do(opData, stateMap, testnet)
        if err != nil {
            return storage.DataRollbackType{}, 0, err
        }
        if opData.OpAccept == 1 {
            cpHeader := strconv.FormatUint(opData.OpScore,10) +","+ opData.TxId +","+ opData.BlockAccept +","+ opData.OpScript.P +","+ opData.OpScript.Op
            sum := blake2b.Sum256([]byte(cpHeader))
            cpHeader = fmt.Sprintf("%064x", string(sum[:]))
            cpState := strings.Join(opData.StAfter, ";")
            sum = blake2b.Sum256([]byte(cpState))
            cpState = fmt.Sprintf("%064x", string(sum[:]))
            sum = blake2b.Sum256([]byte(checkpointLast + cpHeader + cpState))
            opData.Checkpoint = fmt.Sprintf("%064x", string(sum[:]))
            checkpointLast = opData.Checkpoint
        }
        rollback.OpScoreLast = opData.OpScore
        rollback.OpScoreList = append(rollback.OpScoreList, opData.OpScore)
        rollback.TxIdList = append(rollback.TxIdList, opData.TxId)
    }
    rollback.CheckpointAfter = checkpointLast
    return rollback, time.Now().UnixMilli() - mtss, nil
}

////////////////////////////////
func MakeStLineToken(key string, stToken *storage.StateTokenType, isDeploy bool) (string) {
    stLine := storage.KeyPrefixStateToken + key
    if stToken == nil {
        return stLine
    }
    stLine += ","
    strDec := strconv.Itoa(stToken.Dec)
    opScore := stToken.OpMod
    if isDeploy {
        opScore = stToken.OpAdd
    }
    strOpscore := strconv.FormatUint(opScore, 10)
    if isDeploy {
        stLine += stToken.Max + ","
        stLine += stToken.Lim + ","
        stLine += stToken.Pre + ","
        stLine += strDec + ","
        stLine += stToken.From + ","
        stLine += stToken.To + ","
    }
    stLine += stToken.Minted + ","
    stLine += strOpscore
    return stLine
}

////////////////////////////////
func MakeStLineBalance(key string, stBalance *storage.StateBalanceType) (string) {
    stLine := storage.KeyPrefixStateBalance + key
    if stBalance == nil {
        return stLine
    }
    stLine += ","
    strDec := strconv.Itoa(stBalance.Dec)
    strOpscore := strconv.FormatUint(stBalance.OpMod, 10)
    stLine += strDec + ","
    stLine += stBalance.Balance + ","
    stLine += stBalance.Locked + ","
    stLine += strOpscore
    return stLine
}

////////////////////////////////
func ValidateTick(tick *string) (bool) {
    *tick = strings.ToUpper(*tick)
    lenTick := len(*tick)
    if (lenTick < 4 || lenTick > 6) {
        return false
    }
    for i := 0; i < lenTick; i++ {
        if ((*tick)[i] < 65 || (*tick)[i] > 90) {
            return false
        }
    }
    return true
}
////////////////////////////////
func ValidateAmount(amount *string) (bool) {
    if *amount == "" {
        *amount = "0"
        return false
    }
    amountBig := new(big.Int)
    _, s := amountBig.SetString(*amount, 10)
    if !s {
        return false
    }
    amount2 := amountBig.Text(10)
    if *amount != amount2 {
        return false
    }
    limitBig := new(big.Int)
    limitBig.SetString("0", 10)
    if limitBig.Cmp(amountBig) >= 0 {
        return false
    }
    limitBig.SetString("99999999999999999999999999999999", 10)
    if amountBig.Cmp(limitBig) > 0 {
        return false
    }
    return true
}

////////////////////////////////
func ValidateDec(dec *string, def string) (bool) {
    if *dec == "" {
        *dec = def
        return true
    }
    decInt, err := strconv.Atoi(*dec)
    if err != nil {
        return false
    }
    decString := strconv.Itoa(decInt)
    if (decString != *dec || decInt < 0 || decInt > 18) {
        return false
    }
    return true
}

////////////////////////////////
func ValidationUint(value *string, def string) (bool) {
    if *value == "" {
        *value = def
        return true
    }
    valueUint, err := strconv.ParseUint(*value, 10, 64)
    if err != nil {
        return false
    }
    valueString := strconv.FormatUint(valueUint, 10)
    if (valueString != *value) {
        return false
    }
    return true
}

// ...
