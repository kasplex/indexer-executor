
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
func (opMethodDeploy OpMethodDeploy) Validate(script *storage.DataScriptType, testnet bool) (bool) {
    if (script.From == "" || script.P != "KRC-20" || !ValidateTick(&script.Tick) || !ValidateAmount(&script.Max) || !ValidateAmount(&script.Lim) || !ValidateDec(&script.Dec, "8")) {
        return false
    }
    ValidateAmount(&script.Pre)
    if script.To == "" {
        script.To = script.From
    }
    return true
}

////////////////////////////////
func (opMethodDeploy OpMethodDeploy) PrepareStateKey(opData *storage.DataOperationType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opData.OpScript.Tick] = nil
    if opData.OpScript.Pre != "0" {
        stateMap.StateBalanceMap[opData.OpScript.To+"_"+opData.OpScript.Tick] = nil
    }
}

////////////////////////////////
func (opMethodDeploy OpMethodDeploy) Do(opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    ////////////////////////////////
    if stateMap.StateTokenMap[opData.OpScript.Tick] != nil {
        opData.OpAccept = -1
        opData.OpError = "tick existed"
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
    if (TickReserved[opData.OpScript.Tick] != "" && TickReserved[opData.OpScript.Tick] != opData.OpScript.From) {
        opData.OpAccept = -1
        opData.OpError = "tick reserved"
        return nil
    }
    if (opData.OpScript.Pre != "0" && !misc.VerifyAddr(opData.OpScript.To, testnet)) {
        opData.OpAccept = -1
        opData.OpError = "address invalid"
        return nil
    }
    ////////////////////////////////
    keyBalance := opData.OpScript.To +"_"+ opData.OpScript.Tick
    stToken := stateMap.StateTokenMap[opData.OpScript.Tick]
    stBalance := stateMap.StateBalanceMap[keyBalance]
    ////////////////////////////////
    opData.StBefore = nil
    stLine := MakeStLineToken(opData.OpScript.Tick, stToken, true)
    opData.StBefore = append(opData.StBefore, stLine)
    if opData.OpScript.Pre != "0" {
        stLine = MakeStLineBalance(keyBalance, stBalance)
        opData.StBefore = append(opData.StBefore, stLine)
    }
    ////////////////////////////////
    decInt, _ := strconv.Atoi(opData.OpScript.Dec)
    minted := "0"
    stToken = &storage.StateTokenType{
        Tick: opData.OpScript.Tick,
        Max: opData.OpScript.Max,
        Lim: opData.OpScript.Lim,
        Pre: opData.OpScript.Pre,
        Dec: decInt,
        From: opData.OpScript.From,
        To: opData.OpScript.To,
        Minted: minted,
        TxId: opData.TxId,
        OpAdd: opData.OpScore,
        OpMod: opData.OpScore,
        MtsAdd: opData.MtsAdd,
        MtsMod: opData.MtsAdd,
    }
    stateMap.StateTokenMap[opData.OpScript.Tick] = stToken
    if opData.OpScript.Pre != "0" {
        minted = opData.OpScript.Pre
        maxBig := new(big.Int)
        maxBig.SetString(opData.OpScript.Max, 10)
        preBig := new(big.Int)
        preBig.SetString(opData.OpScript.Pre, 10)
        if preBig.Cmp(maxBig) > 0 {
            minted = opData.OpScript.Max
        }
        stToken.Minted = minted
        stBalance = &storage.StateBalanceType{
            Address: opData.OpScript.To,
            Tick: opData.OpScript.Tick,
            Dec: decInt,
            Balance: minted,
            Locked: "0",
            OpMod: opData.OpScore,
        }
        stateMap.StateBalanceMap[keyBalance] = stBalance
        ////////////////////////////
        opData.SsInfo.TickAffc = append(opData.SsInfo.TickAffc, opData.OpScript.Tick+"=1")
        opData.SsInfo.AddressAffc = append(opData.SsInfo.AddressAffc, opData.OpScript.To+"_"+opData.OpScript.Tick+"="+minted)
    } else {
        ////////////////////////////
        opData.SsInfo.TickAffc = append(opData.SsInfo.TickAffc, opData.OpScript.Tick+"=0")
    }
    ////////////////////////////////
    opData.StAfter = nil
    stLine = MakeStLineToken(opData.OpScript.Tick, stToken, true)
    opData.StAfter = append(opData.StAfter, stLine)
    if opData.OpScript.Pre != "0" {
        stLine = MakeStLineBalance(keyBalance, stBalance)
        opData.StAfter = append(opData.StAfter, stLine)
    }
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
func (opMethodDeploy OpMethodDeploy) UnDo() (error) {
    // ...
    return nil
}

// ...
