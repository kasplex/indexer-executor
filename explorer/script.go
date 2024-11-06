
////////////////////////////////
package explorer

import (
    "time"
    "sync"
    "strconv"
    "strings"
    "unicode"
    //"log/slog"
    "encoding/hex"
    "encoding/json"
    "kasplex-executor/misc"
    "kasplex-executor/storage"
    "kasplex-executor/operation"
)

////////////////////////////////
// Parse the P2SH transaction input script.
func parseScriptInput(script string) (bool, []string) {
    script = strings.ToLower(script)
    lenScript := len(script)
    if (lenScript <= 138) {
        return false, nil
    }
    // Get the next data length and position.
    _lGet := func(s string, i int) (int64, int, bool) {
        iRaw := i
        lenS := len(s)
        if lenS < (i + 2) {
            return 0, iRaw, false
        }
        f := s[i:i+2]
        i += 2
        lenD := int64(0)
        if f == "4c" {
            if lenS < (i + 2) {
                return 0, iRaw, false
            }
            f := s[i:i+2]
            i += 2
            lenD, _ = strconv.ParseInt(f, 16, 32)
        } else if f == "4d" {
            if lenS < (i + 4) {
                return 0, iRaw, false
            }
            f := s[i+2:i+4] + s[i:i+2]
            i += 4
            lenD, _ = strconv.ParseInt(f, 16, 32)
        } else {
            lenD, _ = strconv.ParseInt(f, 16, 32)
            if (lenD <0 || lenD > 75) {
                return 0, iRaw, false
            }
        }
        lenD *= 2
        return lenD, i, true
    }
    
    // Get the push number and position.
    _nGet := func(s string, i int) (int64, int, bool) {
        iRaw := i
        lenS := len(s)
        if lenS < (i + 2) {
            return 0, iRaw, false
        }
        f := s[i:i+2]
        i += 2
        num, _ := strconv.ParseInt(f, 16, 32)
        if (num < 81 || num > 96) {
            return 0, iRaw, false
        }
        num -= 80
        return num, i, true
    }
    
    // Get the last data position.
    _dGotoLast := func(s string, i int) (int, bool) {
        iRaw := i
        lenS := len(s)
        lenD := int64(0)
        r := true
        for j := 0; j < 16; j ++ {
            lenD, i, r = _lGet(s, i)
            if !r {
                return iRaw, false
            }
            if lenS < (i + int(lenD)) {
                return iRaw, false
            } else if lenS == (i + int(lenD)) {
                if lenD < 94 {
                    return iRaw, false
                }
                return i, true
            } else {
                i += int(lenD)
            }
        }
        return iRaw, false
    }
    
    // Skip to the redeem script.
    r := true
    n := 0
    flag := ""
    n, r = _dGotoLast(script, n)
    if !r {
        return false, nil
    }
    
    // Get the public key or multisig script hash
    multisig := false
    mm := int64(0)
    nn := int64(0)
    kPub := ""
    lenD := int64(0)
    mm, n, r = _nGet(script, n)
    if r {
        if (mm > 0 && mm < 16) {
            multisig = true
        } else {
            return false, nil
        }
    }
    if !multisig {
        lenD, n, r = _lGet(script, n)
        if !r {
            return false, nil
        }
        fSig := ""
        if lenScript > (n + int(lenD) + 2) {
            fSig = script[n+int(lenD):n+int(lenD)+2]
        }
        if (lenD == 64 && fSig == "ac") {
            kPub = script[n:n+64]
            n += 66
        } else if (lenD == 66 && fSig == "ab") {
            kPub = script[n:n+66]
            n += 68
        } else {
            return false, nil
        }
    } else {
        var kPubList []string
        for j := 0; j < 16; j ++ {
            lenD, n, r = _lGet(script, n)
            if !r {
                nn, n, r = _nGet(script, n)
                if (!r || len(kPubList) != int(nn)) {
                    return false, nil
                }
                kPub = misc.ConvKPubListToScriptHashMultisig(mm, kPubList, nn)
                break
            }
            if (lenD == 64 || lenD == 66) {
                kPubList = append(kPubList, script[n:n+int(lenD)])
                n += int(lenD)
            } else {
                return false, nil
            }
        }
        if lenScript < (n + 2) {
            return false, nil
        }
        flag = script[n:n+2]
        n += 2
        if (flag != "a9" && flag != "ae") {
            return false, nil
        }
    }
    if kPub == "" {
        return false, nil
    }
    // Check the protocol header.
    if lenScript < (n + 22) {
        return false, nil
    }
    flag = script[n:n+6]
    n += 6
    if flag != "006307" {
        return false, nil
    }
    flag = script[n:n+14]
    n += 14
    decoded, _ := hex.DecodeString(flag)
    header := strings.ToUpper(string(decoded[:]))
    if header != "KASPLEX" {
        return false, nil
    }
    
    // Get the next param data and position.
    _pGet := func(s string, i int) (string, int, bool) {
        iRaw := i
        lenS := len(s)
        lenP := int64(0)
        lenP, i, r = _lGet(s, i)
        if (!r || lenS < (i + int(lenP))) {
            return "", iRaw, false
        }
        if lenP == 0 {
            return "", i, true
        }
        decoded, _ = hex.DecodeString(s[i:i+int(lenP)])
        p := string(decoded[:])
        i += int(lenP)
        return p, i, true
    }
    
    // Get the param and json data.
    p0 := ""
    p1 := ""
    p2 := ""
    r = true
    for j := 0; j < 2; j ++ {
        if lenScript < (n + 2) {
            return false, nil
        }
        flag = script[n:n+2]
        n += 2
        if flag == "00" {
            p0, n, r = _pGet(script, n)
        } else if flag == "68" {
            break
        } else {
            if flag == "51" {
                p1 = "p1"
            } else if flag == "53" {
                p1 = "p3"
            } else if flag == "55" {
                p1 = "p5"
            } else if flag == "57" {
                p1 = "p7"
            } else if flag == "59" {
                p1 = "p9"
            } else if flag == "5b" {
                p1 = "p11"
            } else if flag == "5d" {
                p1 = "p13"
            } else if flag == "5f" {
                p1 = "p15"
            } else {
                return false, nil
            }
            p2, n, r = _pGet(script, n)
        }
        if !r {
            return false, nil
        }
    }
    if p0 == "" {
        return false, nil
    }
    
    // Get the from address.
    from := ""
    if multisig {
        from = misc.ConvKPubToP2sh(kPub, eRuntime.testnet)
    } else {
        from = misc.ConvKPubToAddr(kPub, eRuntime.testnet)
    }
    return true, []string{from, p0, p1, p2}
}

////////////////////////////////
// Parse the OP data in transaction.
func parseOpData(txData *storage.DataTransactionType) (*storage.DataOperationType, error) {
    if (txData == nil || txData.Data == nil) {
        return nil, nil
    }
    lenInput := len(txData.Data.Inputs)
    if lenInput <= 0 {
        return nil, nil
    }
    script := txData.Data.Inputs[0].SignatureScript
    isOp, scriptInfo := parseScriptInput(script)
    if !isOp {
        return nil, nil
    }
    if scriptInfo[0] == "" {
        return nil, nil
    }
    decoded := storage.DataScriptType{}
    err := json.Unmarshal([]byte(scriptInfo[1]), &decoded)
    if err != nil {
        return nil, nil
    }
    decoded.From = scriptInfo[0]
    if (!eRuntime.testnet && txData.DaaScore <= 83525600) {  // use output[0] as the to-address
        decoded.To = txData.Data.Outputs[0].VerboseData.ScriptPublicKeyAddress
    }
    if (!ValidateP(&decoded.P) || !ValidateOp(&decoded.Op) || !ValidateAscii(&decoded.To)) {
        return nil, nil
    }
    if !operation.Method_Registered[decoded.Op].Validate(&decoded, eRuntime.testnet) {
        return nil, nil
    }
    opData := &storage.DataOperationType{
        TxId: txData.TxId,
        DaaScore: txData.DaaScore,
        BlockAccept: txData.BlockAccept,
        MtsAdd: int64(txData.Data.VerboseData.BlockTime),
        OpScript: &decoded,
        SsInfo: &storage.DataStatsType{},
    }
    return opData, nil
}

////////////////////////////////
// Parse the OP data in transaction list.
func ParseOpDataList(txDataList []storage.DataTransactionType) ([]storage.DataOperationType, int64, error) {
    mtss := time.Now().UnixMilli()
    opDataMap := map[string]*storage.DataOperationType{}
    txIdMap := map[string]bool{}
    mutex := new(sync.RWMutex)
    misc.GoBatch(len(txDataList), func(i int) (error) {
        opData, err := parseOpData(&txDataList[i])
        if err != nil {
            return err
        }
        if opData == nil {
            return nil
        }
        mutex.Lock()
        opDataMap[opData.TxId] = opData
        opDataMap[opData.TxId].FeeLeast = operation.Method_Registered[opData.OpScript.Op].FeeLeast(opData.DaaScore)
        if opDataMap[opData.TxId].FeeLeast > 0 {
            for _, input := range txDataList[i].Data.Inputs {
                txIdMap[input.PreviousOutpoint.TransactionId] = true
            }
        }
        mutex.Unlock()
        return nil
    })
    txDataListInput := make([]storage.DataTransactionType, 0, len(txIdMap))
    for txId := range txIdMap {
        txDataListInput = append(txDataListInput, storage.DataTransactionType{TxId: txId})
    }
    txDataMapInput, _, err := storage.GetNodeTransactionDataMap(txDataListInput)
    if err != nil {
        return nil, 0, err
    }
    opDataList := []storage.DataOperationType{}
    daaScoreNow := uint64(0)
    opScore := uint64(0)
    for _, txData := range txDataList {
        if opDataMap[txData.TxId] == nil {
            continue
        }
        if daaScoreNow != txData.DaaScore {
            daaScoreNow = txData.DaaScore
            opScore = daaScoreNow * 10000
        }
        opDataMap[txData.TxId].OpScore = opScore
        if opDataMap[txData.TxId].FeeLeast > 0 {
            amountIn := uint64(0)
            amountOut := uint64(0)
            for _, output := range txData.Data.Outputs {
                amountOut += output.Amount
            }
            for _, input := range txData.Data.Inputs {
                if txDataMapInput[input.PreviousOutpoint.TransactionId] == nil {
                    continue
                }
                amountIn += txDataMapInput[input.PreviousOutpoint.TransactionId].Outputs[input.PreviousOutpoint.Index].Amount
            }
            if amountIn <= amountOut {
                opDataMap[txData.TxId].Fee = 0
                continue
            }
            opDataMap[txData.TxId].Fee = amountIn - amountOut
        }
        opDataList = append(opDataList, *opDataMap[txData.TxId])
        opScore ++
    }
    return opDataList, time.Now().UnixMilli() - mtss, nil
}

////////////////////////////////
func ValidateP(p *string) (bool) {
    *p = strings.ToUpper(*p)
    if !operation.P_Registered[*p] {
        return false
    }
    return true
}

////////////////////////////////
func ValidateOp(op *string) (bool) {
    *op = strings.ToLower(*op)
    if !operation.Op_Registered[*op] {
        return false
    }
    return true
}

////////////////////////////////
func ValidateAscii(s *string) (bool) {
    if *s == "" {
        return true
    }
    for _, c := range *s {
        if c > unicode.MaxASCII {
            return false
        }
    }
    return true
}
