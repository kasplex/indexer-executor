
////////////////////////////////
package operation

import (
    "math/big"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethodBurn struct {}

////////////////////////////////
func init() {
    opName := "burn"
    P_Registered["KRC-20"] = true
    Op_Registered[opName] = true
    Method_Registered[opName] = new(OpMethodBurn)
}

////////////////////////////////
func (opMethodBurn OpMethodBurn) FeeLeast(daaScore uint64) (uint64) {
    // if daaScore ...
    return 0
}

////////////////////////////////
func (opMethodBurn OpMethodBurn) ScriptCollectEx(index int, script *storage.DataScriptType, txData *storage.DataTransactionType, testnet bool) {}

////////////////////////////////
func (opMethodBurn OpMethodBurn) Validate(script *storage.DataScriptType, txId string, daaScore uint64, testnet bool) (bool) {
    if (!testnet && daaScore < 9999999999) {  // undetermined for mainnet
        return false
    }
    if ValidateTxId(&script.Sc) {
        script.Tick = script.Sc
    }
    if (script.From == "" || script.P != "KRC-20" || !ValidateTickTxId(&script.Tick) || !ValidateAmount(&script.Amt)) {
        return false
    }
    script.To = ""
    script.Max = ""
    script.Lim = ""
    script.Pre = ""
    script.Dec = ""
    script.Utxo = ""
    script.Price = ""
    script.Mod = ""
    script.Name = ""
    script.Sc = ""
    return true
}

////////////////////////////////
func (opMethodBurn OpMethodBurn) PrepareStateKey(opScript *storage.DataScriptType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opScript.Tick] = nil
    stateMap.StateBalanceMap[opScript.From+"_"+opScript.Tick] = nil
}

////////////////////////////////
func (opMethodBurn OpMethodBurn) Do(index int, opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    opScript := opData.OpScript[index]
    ////////////////////////////////
    if stateMap.StateTokenMap[opScript.Tick] == nil {
        opData.OpAccept = -1
        opData.OpError = "tick not found"
        return nil
    }
    if opScript.From != stateMap.StateTokenMap[opScript.Tick].To {
        opData.OpAccept = -1
        opData.OpError = "no ownership"
        return nil
    }
    ////////////////////////////////
    opScript.Name = stateMap.StateTokenMap[opScript.Tick].Name
    keyBalance := opScript.From +"_"+ opScript.Tick
    stToken := stateMap.StateTokenMap[opScript.Tick]
    stBalance := stateMap.StateBalanceMap[keyBalance]
    nTickAffc := int64(0)
    opScript.Name = stToken.Name
    ////////////////////////////////
    if stBalance == nil {
        opData.OpAccept = -1
        opData.OpError = "balance insuff"
        return nil
    }
    balanceBig := new(big.Int)
    balanceBig.SetString(stBalance.Balance, 10)
    amtBig := new(big.Int)
    amtBig.SetString(opScript.Amt, 10)
    if amtBig.Cmp(balanceBig) > 0 {
        opData.OpAccept = -1
        opData.OpError = "balance insuff"
        return nil
    } else if (amtBig.Cmp(balanceBig) == 0 && stBalance.Locked == "0") {
        nTickAffc = -1
    }
    ////////////////////////////////
    opData.StBefore = nil
    opData.StBefore = AppendStLineToken(opData.StBefore, opScript.Tick, stToken, false, false)
    opData.StBefore = AppendStLineBalance(opData.StBefore, keyBalance, stBalance, false)
    ////////////////////////////////
    balanceBig = balanceBig.Sub(balanceBig, amtBig)
    stBalance.Balance = balanceBig.Text(10)
    stBalance.OpMod = opData.OpScore
    burnedBig := new(big.Int)
    burnedBig.SetString(stToken.Burned, 10)
    burnedBig = burnedBig.Add(burnedBig, amtBig)
    stToken.Burned = burnedBig.Text(10)
    stToken.OpMod = opData.OpScore
    stToken.MtsMod = opData.MtsAdd
    ////////////////////////////////
    opData.SsInfo.TickAffc = AppendSsInfoTickAffc(opData.SsInfo.TickAffc, opScript.Tick, nTickAffc)
    lockedBig := new(big.Int)
    lockedBig.SetString(stBalance.Locked, 10)
    balanceBig = balanceBig.Add(balanceBig, lockedBig)
    balanceTotal := balanceBig.Text(10)
    opData.SsInfo.AddressAffc = AppendSsInfoAddressAffc(opData.SsInfo.AddressAffc, keyBalance, balanceTotal)
    ////////////////////////////////
    opData.StAfter = nil
    opData.StAfter = AppendStLineToken(opData.StAfter, opScript.Tick, stToken, false, true)
    opData.StAfter = AppendStLineBalance(opData.StAfter, keyBalance, stBalance, true)
    ////////////////////////////////
    if (stBalance.Balance == "0" && stBalance.Locked == "0") {
        stateMap.StateBalanceMap[keyBalance] = nil
    }
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
/*func (opMethodBurn OpMethodBurn) UnDo() (error) {
    // ...
    return nil
}*/

// ...
