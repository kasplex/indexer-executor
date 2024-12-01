
////////////////////////////////
package operation

import (
    "strconv"
    "math/big"
    "kasplex-executor/misc"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethodDeploy struct {}

////////////////////////////////
func init() {
    opName := "deploy"
    P_Registered["KRC-20"] = true
    Op_Registered[opName] = true
    Method_Registered[opName] = new(OpMethodDeploy)
}

////////////////////////////////
func (opMethodDeploy OpMethodDeploy) FeeLeast(daaScore uint64) (uint64) {
    // if daaScore ...
    return 100000000000
}

////////////////////////////////
func (opMethodDeploy OpMethodDeploy) ScriptCollectEx(index int, script *storage.DataScriptType, txData *storage.DataTransactionType, testnet bool) {}

////////////////////////////////
func (opMethodDeploy OpMethodDeploy) Validate(script *storage.DataScriptType, daaScore uint64, testnet bool) (bool) {
    if (script.From == "" || script.P != "KRC-20" || !ValidateTick(&script.Tick) || !ValidateAmount(&script.Max) || !ValidateAmount(&script.Lim) || !ValidateDec(&script.Dec, "8")) {
        return false
    }
    if !ValidateAmount(&script.Pre) {
        script.Pre = "0"
    }
    if script.To == "" {
        script.To = script.From
    }
    script.Amt = ""
    script.Utxo = ""
    script.Price = ""
    return true
}

////////////////////////////////
func (opMethodDeploy OpMethodDeploy) PrepareStateKey(opScript *storage.DataScriptType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opScript.Tick] = nil
    if opScript.Pre != "0" {
        stateMap.StateBalanceMap[opScript.To+"_"+opScript.Tick] = nil
    }
}

////////////////////////////////
func (opMethodDeploy OpMethodDeploy) Do(index int, opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    opScript := opData.OpScript[index]
    ////////////////////////////////
    if stateMap.StateTokenMap[opScript.Tick] != nil {
        opData.OpAccept = -1
        opData.OpError = "tick existed"
        return nil
    }
    if TickIgnored[opScript.Tick] {
        opData.OpAccept = -1
        opData.OpError = "tick ignored"
        return nil
    }
    if (TickReserved[opScript.Tick] != "" && TickReserved[opScript.Tick] != opScript.From) {
        opData.OpAccept = -1
        opData.OpError = "tick reserved"
        return nil
    }
    if opData.Fee == 0 {
        opData.OpAccept = -1
        opData.OpError = "fee unknown"
        return nil
    }
    if opData.Fee < opData.FeeLeast {
        opData.OpAccept = -1
        opData.OpError = "fee not enough"
        return nil
    }
    if (opScript.Pre != "0" && !misc.VerifyAddr(opScript.To, testnet)) {
        opData.OpAccept = -1
        opData.OpError = "address invalid"
        return nil
    }
    ////////////////////////////////
    keyBalance := opScript.To +"_"+ opScript.Tick
    stToken := stateMap.StateTokenMap[opScript.Tick]
    stBalance := stateMap.StateBalanceMap[keyBalance]
    ////////////////////////////////
    opData.StBefore = nil
    opData.StBefore = AppendStLineToken(opData.StBefore, opScript.Tick, stToken, true, false)
    if opScript.Pre != "0" {
        opData.StBefore = AppendStLineBalance(opData.StBefore, keyBalance, stBalance, false)
    }
    ////////////////////////////////
    decInt, _ := strconv.Atoi(opScript.Dec)
    minted := "0"
    stToken = &storage.StateTokenType{
        Tick: opScript.Tick,
        Max: opScript.Max,
        Lim: opScript.Lim,
        Pre: opScript.Pre,
        Dec: decInt,
        From: opScript.From,
        To: opScript.To,
        Minted: minted,
        TxId: opData.TxId,
        OpAdd: opData.OpScore,
        OpMod: opData.OpScore,
        MtsAdd: opData.MtsAdd,
        MtsMod: opData.MtsAdd,
    }
    stateMap.StateTokenMap[opScript.Tick] = stToken
    if opScript.Pre != "0" {
        minted = opScript.Pre
        maxBig := new(big.Int)
        maxBig.SetString(opScript.Max, 10)
        preBig := new(big.Int)
        preBig.SetString(opScript.Pre, 10)
        if preBig.Cmp(maxBig) > 0 {
            minted = opScript.Max
        }
        stToken.Minted = minted
        stBalance = &storage.StateBalanceType{
            Address: opScript.To,
            Tick: opScript.Tick,
            Dec: decInt,
            Balance: minted,
            Locked: "0",
            OpMod: opData.OpScore,
        }
        stateMap.StateBalanceMap[keyBalance] = stBalance
        ////////////////////////////
        opData.SsInfo.TickAffc = AppendSsInfoTickAffc(opData.SsInfo.TickAffc, opScript.Tick, 1)
        opData.SsInfo.AddressAffc = AppendSsInfoAddressAffc(opData.SsInfo.AddressAffc, opScript.To+"_"+opScript.Tick, minted)
    } else {
        ////////////////////////////
        opData.SsInfo.TickAffc = AppendSsInfoTickAffc(opData.SsInfo.TickAffc, opScript.Tick, 0)
    }
    ////////////////////////////////
    opData.StAfter = nil
    opData.StAfter = AppendStLineToken(opData.StAfter, opScript.Tick, stToken, true, true)
    if opScript.Pre != "0" {
        opData.StAfter = AppendStLineBalance(opData.StAfter, keyBalance, stBalance, true)
    }
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
/*func (opMethodDeploy OpMethodDeploy) UnDo() (error) {
    // ...
    return nil
}*/

// ...
