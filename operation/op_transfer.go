
////////////////////////////////
package operation

import (
    //"strconv"
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
func (opMethodTransfer OpMethodTransfer) ScriptCollectEx(index int, script *storage.DataScriptType, txData *storage.DataTransactionType, testnet bool) {}

////////////////////////////////
func (opMethodTransfer OpMethodTransfer) Validate(script *storage.DataScriptType, daaScore uint64, testnet bool) (bool) {
    if (script.From == "" || script.To == "" || script.P != "KRC-20" || !ValidateTick(&script.Tick) || !ValidateAmount(&script.Amt)) {
        return false
    }
    script.Max = ""
    script.Lim = ""
    script.Pre = ""
    script.Dec = ""
    script.Utxo = ""
    script.Price = ""
    return true
}

////////////////////////////////
func (opMethodTransfer OpMethodTransfer) PrepareStateKey(opScript *storage.DataScriptType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opScript.Tick] = nil
    stateMap.StateBalanceMap[opScript.From+"_"+opScript.Tick] = nil
    stateMap.StateBalanceMap[opScript.To+"_"+opScript.Tick] = nil
}

////////////////////////////////
func (opMethodTransfer OpMethodTransfer) Do(index int, opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    opScript := opData.OpScript[index]
    ////////////////////////////////
    if stateMap.StateTokenMap[opScript.Tick] == nil {
        opData.OpAccept = -1
        opData.OpError = "tick not found"
        return nil
    }
    if (opScript.From == opScript.To || !misc.VerifyAddr(opScript.To, testnet)) {
        opData.OpAccept = -1
        opData.OpError = "address invalid"
        return nil
    }
    ////////////////////////////////
    keyBalanceFrom := opScript.From +"_"+ opScript.Tick
    keyBalanceTo := opScript.To +"_"+ opScript.Tick
    stBalanceFrom := stateMap.StateBalanceMap[keyBalanceFrom]
    stBalanceTo := stateMap.StateBalanceMap[keyBalanceTo]
    nTickAffc := int64(0)
    ////////////////////////////////
    if stBalanceFrom == nil {
        opData.OpAccept = -1
        opData.OpError = "balance insuff"
        return nil
    }
    balanceBig := new(big.Int)
    balanceBig.SetString(stBalanceFrom.Balance, 10)
    amtBig := new(big.Int)
    amtBig.SetString(opScript.Amt, 10)
    if amtBig.Cmp(balanceBig) > 0 {
        opData.OpAccept = -1
        opData.OpError = "balance insuff"
        return nil
    } else if (amtBig.Cmp(balanceBig) == 0 && stBalanceFrom.Locked == "0") {
        nTickAffc = -1
    }
    ////////////////////////////////
    opData.StBefore = nil
    opData.StBefore = AppendStLineBalance(opData.StBefore, keyBalanceFrom, stBalanceFrom, false)
    opData.StBefore = AppendStLineBalance(opData.StBefore, keyBalanceTo, stBalanceTo, false)
    ////////////////////////////////
    balanceBig = balanceBig.Sub(balanceBig, amtBig)
    stBalanceFrom.Balance = balanceBig.Text(10)
    stBalanceFrom.OpMod = opData.OpScore
    lockedBig := new(big.Int)
    lockedBig.SetString(stBalanceFrom.Locked, 10)
    balanceBig = balanceBig.Add(balanceBig, lockedBig)
    balanceFromTotal := balanceBig.Text(10)
    if stBalanceTo == nil {
        stBalanceTo = &storage.StateBalanceType{
            Address: opScript.To,
            Tick: opScript.Tick,
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
    lockedBig.SetString(stBalanceTo.Locked, 10)
    balanceBig = balanceBig.Add(balanceBig, lockedBig)
    balanceToTotal := balanceBig.Text(10)
    ////////////////////////////////
    opData.SsInfo.TickAffc = AppendSsInfoTickAffc(opData.SsInfo.TickAffc, opScript.Tick, nTickAffc)
    opData.SsInfo.AddressAffc = AppendSsInfoAddressAffc(opData.SsInfo.AddressAffc, opScript.From+"_"+opScript.Tick, balanceFromTotal)
    opData.SsInfo.AddressAffc = AppendSsInfoAddressAffc(opData.SsInfo.AddressAffc, opScript.To+"_"+opScript.Tick, balanceToTotal)
    ////////////////////////////////
    opData.StAfter = nil
    opData.StAfter = AppendStLineBalance(opData.StAfter, keyBalanceFrom, stBalanceFrom, true)
    opData.StAfter = AppendStLineBalance(opData.StAfter, keyBalanceTo, stBalanceTo, true)
    ////////////////////////////////
    if (stBalanceFrom.Balance == "0" && stBalanceFrom.Locked == "0") {
        stateMap.StateBalanceMap[keyBalanceFrom] = nil
    }
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
/*func (opMethodTransfer OpMethodTransfer) UnDo() (error) {
    // ...
    return nil
}*/

// ...
