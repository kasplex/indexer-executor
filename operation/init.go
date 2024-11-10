
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
    ScriptCollectEx(int, *storage.DataScriptType, *storage.DataTransactionType, bool)
    Validate(*storage.DataScriptType, uint64, bool) (bool)
    FeeLeast(uint64) (uint64)
    PrepareStateKey(*storage.DataScriptType, storage.DataStateMapType)
    Do(int, *storage.DataOperationType, storage.DataStateMapType, bool) (error)
    //UnDo() (error)
    // ...
}

////////////////////////////////
var P_Registered = map[string]bool{}
var Op_Registered = map[string]bool{}
var Method_Registered = map[string]OpMethod{}
var OpRecycle_Registered = map[string]bool{}

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
var TickReserved = map[string]string{
    "NACHO": "kaspa:qzrsq2mfj9sf7uye3u5q7juejzlr0axk5jz9fpg4vqe76erdyvxxze84k9nk7",
    "KCATS": "kaspa:qq8guq855gxkfrj2w25skwgj7cp4hy08x6a8mz70tdtmgv5p2ngwqxpj4cknc",
    "KASTOR": "kaspa:qr8vt54764aaddejhjfwtsh07jcjr49v38vrw2vtmxxtle7j2uepynwy57ufg",
    "KASPER": "kaspa:qppklkx2zyr2g2djg3uy2y2tsufwsqjk36pt27vt2xfu8uqm24pskk4p7tq5n",
    "FUSUN": "kaspa:qzp30gu5uty8jahu9lq5vtplw2ca8m2k7p45ez3y8jf9yrm5qdxquq5nl45t5",
    "KPAW": "kaspa:qpp0y685frmnlvhmnz5t6qljatumqm9zmppwnhwu9vyyl6w8nt30qjedekmdw",
    "PPKAS": "kaspa:qrlx9377yje3gvj9qxvwnn697d209lshgcrvge3yzlxnvyrfyk3q583jh3cmz",
    "GHOAD": "kaspa:qpkty3ymqs67t0z3g7l457l79f9k6drl55uf2qeq5tlkrpf3zwh85es0xtaj9",
    "KEPE": "kaspa:qq45gur2grn80uuegg9qgewl0wg2ahz5n4qm9246laej9533f8e22x3xe6hkm",
    "WORI": "kaspa:qzhgepc7mjscszkteeqhy99d3v96ftpg2wyy6r85nd0kg9m8rfmusqpp7mxkq",
    "KEKE": "kaspa:qqq9m42mdcvlz8c7r9kmpqj59wkfx3nppqte8ay20m4p46x3z0lsyzz34h8uf",
    "DOGK": "kaspa:qpsj64nxtlwceq4e7jvrsrkl0y6dayfyrqr49pep7pd2tq2uzvk7ks7n0qwxc",
    "BTAI": "kaspa:qp0na29g4lysnaep5pmg9xkdzcn4xm4a35ha5naq79ns9mcgc3pccnf225qma",
    "KASBOT": "kaspa:qrrcpdaev9augqwy8jnnp20skplyswa7ezz3m9ex3ryxw22frpzpj2xx99scq",
    "SOMPS": "kaspa:qry7xqy6s7d449gqyl0dkr99x6df0q5jlj6u52p84tfv6rddxjrucnn066237",
    "KREP": "kaspa:qzaclsmr5vttzlt0rz0x3shnudny8lnz5zpmjr4lp9v7aa7u7zvexh05eqwq0",
    // ...
}

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
        StateMarketMap: make(map[string]*storage.StateMarketType),
        // StateXxx ...
    }
    for _, opData := range opDataList{
        for _, opScript := range opData.OpScript{
            Method_Registered[opScript.Op].PrepareStateKey(opScript, stateMap)
        }
    }
    _, err := storage.GetStateTokenMap(stateMap.StateTokenMap)
    if err != nil {
        return storage.DataStateMapType{}, 0, err
    }
    _, err = storage.GetStateBalanceMap(stateMap.StateBalanceMap)
    if err != nil {
        return storage.DataStateMapType{}, 0, err
    }
    _, err = storage.GetStateMarketMap(stateMap.StateMarketMap)
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
        if (testnet && opData.DaaScore%100000 <= 9) {
            checkpointLast = ""
        }
        iScriptAccept := -1
        opError := ""
        for iScript, opScript := range opData.OpScript{
            opData.OpAccept = 0
            opData.OpError = ""
            err := Method_Registered[opScript.Op].Do(iScript, opData, stateMap, testnet)
            if err != nil {
                return storage.DataRollbackType{}, 0, err
            }
            if (opData.OpAccept == 1 && iScriptAccept < 0) {
                iScriptAccept = iScript
            }
            if (opData.OpAccept == -1 && opError == "") {
                opError = opData.OpError
            }
        }
        if iScriptAccept >= 0 {
            opData.OpAccept = 1
            opData.OpError = ""
            if iScriptAccept > 0 {
                opData.OpScript = opData.OpScript[iScriptAccept:]
            }
        } else {
            opData.OpAccept = -1
            opData.OpError = opError
        }
        if opData.OpAccept == 1 {
            cpHeader := strconv.FormatUint(opData.OpScore,10) +","+ opData.TxId +","+ opData.BlockAccept +","+ opData.OpScript[0].P +","+ opData.OpScript[0].Op
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
func AppendStLineToken(stLine []string, key string, stToken *storage.StateTokenType, isDeploy bool, isAfter bool) ([]string) {
    keyFull := storage.KeyPrefixStateToken + key
    iExists := -1
    list := []string{}
    for i, line := range stLine {
        list = strings.SplitN(line, ",", 2)
        if list[0] == keyFull {
            iExists = i
            break
        }
    }
    if iExists < 0 {
        return append(stLine, MakeStLineToken(key, stToken, isDeploy))
    }
    if isAfter {
        stLine[iExists] = MakeStLineToken(key, stToken, isDeploy)
    }
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
func AppendStLineBalance(stLine []string, key string, stBalance *storage.StateBalanceType, isAfter bool) ([]string) {
    keyFull := storage.KeyPrefixStateBalance + key
    iExists := -1
    list := []string{}
    for i, line := range stLine {
        list = strings.SplitN(line, ",", 2)
        if list[0] == keyFull {
            iExists = i
            break
        }
    }
    if iExists < 0 {
        return append(stLine, MakeStLineBalance(key, stBalance))
    }
    if isAfter {
        stLine[iExists] = MakeStLineBalance(key, stBalance)
    }
    return stLine
}

////////////////////////////////
func MakeStLineMarket(key string, stMarket *storage.StateMarketType) (string) {
    stLine := storage.KeyPrefixStateMarket + key
    if stMarket == nil {
        return stLine
    }
    stLine += ","
    strOpscore := strconv.FormatUint(stMarket.OpAdd, 10)
    stLine += stMarket.UAddr + ","
    stLine += stMarket.UAmt + ","
    stLine += stMarket.TAmt + ","
    stLine += strOpscore
    return stLine
}
func AppendStLineMarket(stLine []string, key string, stMarket *storage.StateMarketType, isAfter bool) ([]string) {
    keyFull := storage.KeyPrefixStateMarket + key
    iExists := -1
    list := []string{}
    for i, line := range stLine {
        list = strings.SplitN(line, ",", 2)
        if list[0] == keyFull {
            iExists = i
            break
        }
    }
    if iExists < 0 {
        return append(stLine, MakeStLineMarket(key, stMarket))
    }
    if isAfter {
        stLine[iExists] = MakeStLineMarket(key, stMarket)
    }
    return stLine
}

////////////////////////////////
func AppendSsInfoTickAffc(tickAffc []string, key string, value int64) ([]string) {
    iExists := -1
    valueBefore := int64(0)
    list := []string{}
    for i, affc := range tickAffc {
        list = strings.SplitN(affc, "=", 2)
        if list[0] == key {
            iExists = i
            if len(list) > 1 {
                valueBefore, _ = strconv.ParseInt(list[1], 10, 64)
            }
            break
        }
    }
    if iExists < 0 {
        return append(tickAffc, key+"="+strconv.FormatInt(value, 10))
    }
    tickAffc[iExists] = key+"="+strconv.FormatInt(value+valueBefore, 10)
    return tickAffc
}

////////////////////////////////
func AppendSsInfoAddressAffc(addressAffc []string, key string, value string) ([]string) {
    iExists := -1
    list := []string{}
    for i, affc := range addressAffc {
        list = strings.SplitN(affc, "=", 2)
        if list[0] == key {
            iExists = i
            break
        }
    }
    if iExists < 0 {
        return append(addressAffc, key+"="+value)
    }
    addressAffc[iExists] = key+"="+value
    return addressAffc
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
