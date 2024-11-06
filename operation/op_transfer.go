
////////////////////////////////
package operation

import (
    "strconv"
    "math/big"
    "kasplex-executor/misc"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethodTransfer struct {}

////////////////////////////////
func init() {
    opName := "transfer"
    P_Registered["KRC-20"] = true
    Op_Registered[opName] = true
    Method_Registered[opName] = new(OpMethodTransfer)
}

////////////////////////////////
func (opMethodTransfer OpMethodTransfer) FeeLeast(daaScore uint64) (uint64) {
    return 0
}

////////////////////////////////
func (opMethodTransfer OpMethodTransfer) Validate(script *storage.DataScriptType, testnet bool) (bool) {
    if (script.From == "" || script.To == "" || script.P != "KRC-20" || !ValidateTick(&script.Tick) || !ValidateAmount(&script.Amt)) {
        return false
    }
    script.Max = ""
    script.Lim = ""
    script.Pre = ""
    script.Dec = ""
    return true
}

////////////////////////////////
func (opMethodTransfer OpMethodTransfer) PrepareStateKey(opData *storage.DataOperationType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opData.OpScript.Tick] = nil
    stateMap.StateBalanceMap[opData.OpScript.From+"_"+opData.OpScript.Tick] = nil
    stateMap.StateBalanceMap[opData.OpScript.To+"_"+opData.OpScript.Tick] = nil
}

////////////////////////////////
func (opMethodTransfer OpMethodTransfer) Do(opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    if stateMap.StateTokenMap[opData.OpScript.Tick] == nil {
        opData.OpAccept = -1
        opData.OpError = "tick not found"
        return nil
    }
    if (opData.OpScript.From == opData.OpScript.To || !misc.VerifyAddr(opData.OpScript.To, testnet)) {
        opData.OpAccept = -1
        opData.OpError = "address invalid"
        return nil
    }
    ////////////////////////////////
    keyBalanceFrom := opData.OpScript.From +"_"+ opData.OpScript.Tick
    keyBalanceTo := opData.OpScript.To +"_"+ opData.OpScript.Tick
    stBalanceFrom := stateMap.StateBalanceMap[keyBalanceFrom]
    stBalanceTo := stateMap.StateBalanceMap[keyBalanceTo]
    nTickAffc := 0
    ////////////////////////////////
    if stBalanceFrom == nil {
        opData.OpAccept = -1
        opData.OpError = "balance insuff"
        return nil
    }
    balanceBig := new(big.Int)
    balanceBig.SetString(stBalanceFrom.Balance, 10)
    amtBig := new(big.Int)
    amtBig.SetString(opData.OpScript.Amt, 10)
    if amtBig.Cmp(balanceBig) > 0 {
        opData.OpAccept = -1
        opData.OpError = "balance insuff"
        return nil
    } else if amtBig.Cmp(balanceBig) == 0 {
        nTickAffc = -1
    }
    ////////////////////////////////
    opData.StBefore = nil
    stLine := MakeStLineBalance(keyBalanceFrom, stBalanceFrom)
    opData.StBefore = append(opData.StBefore, stLine)
    stLine = MakeStLineBalance(keyBalanceTo, stBalanceTo)
    opData.StBefore = append(opData.StBefore, stLine)
    ////////////////////////////////
    balanceBig = balanceBig.Sub(balanceBig, amtBig)
    stBalanceFrom.Balance = balanceBig.Text(10)
    stBalanceFrom.OpMod = opData.OpScore
    if stBalanceTo == nil {
        stBalanceTo = &storage.StateBalanceType{
            Address: opData.OpScript.To,
            Tick: opData.OpScript.Tick,
            Dec: stBalanceFrom.Dec,
            Balance: "0",
            Locked: "0",
            OpMod: opData.OpScore,
        }
        stateMap.StateBalanceMap[keyBalanceTo] = stBalanceTo
        nTickAffc ++
    }
    balanceBig.SetString(stBalanceTo.Balance, 10)
    balanceBig = balanceBig.Add(balanceBig, amtBig)
    stBalanceTo.Balance = balanceBig.Text(10)
    stBalanceTo.OpMod = opData.OpScore
    ////////////////////////////////
    opData.SsInfo.TickAffc = append(opData.SsInfo.TickAffc, opData.OpScript.Tick+"="+strconv.Itoa(nTickAffc))
    opData.SsInfo.AddressAffc = append(opData.SsInfo.AddressAffc, opData.OpScript.From+"_"+opData.OpScript.Tick+"="+stBalanceFrom.Balance)
    opData.SsInfo.AddressAffc = append(opData.SsInfo.AddressAffc, opData.OpScript.To+"_"+opData.OpScript.Tick+"="+stBalanceTo.Balance)
    ////////////////////////////////
    opData.StAfter = nil
    stLine = MakeStLineBalance(keyBalanceFrom, stBalanceFrom)
    opData.StAfter = append(opData.StAfter, stLine)
    stLine = MakeStLineBalance(keyBalanceTo, stBalanceTo)
    opData.StAfter = append(opData.StAfter, stLine)
    ////////////////////////////////
    if stBalanceFrom.Balance == "0" {
        stateMap.StateBalanceMap[keyBalanceFrom] = nil
    }
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
func (opMethodTransfer OpMethodTransfer) UnDo() (error) {
    // ...
    return nil
}

// ...
