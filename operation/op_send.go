
////////////////////////////////
package operation

import (
    "strings"
    "strconv"
    "math/big"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethodSend struct {}

////////////////////////////////
func init() {
    opName := "send"
    P_Registered["KRC-20"] = true
    Op_Registered[opName] = true
    OpRecycle_Registered[opName] = true
    Method_Registered[opName] = new(OpMethodSend)
}

////////////////////////////////
func (opMethodSend OpMethodSend) FeeLeast(daaScore uint64) (uint64) {
    return 0
}

////////////////////////////////
func (opMethodSend OpMethodSend) ScriptCollectEx(index int, script *storage.DataScriptType, txData *storage.DataTransactionType, testnet bool) {
    script.Utxo = ""
    script.Price = ""
    script.Utxo = txData.Data.Inputs[index].PreviousOutpoint.TransactionId + "_" + script.From
    if len(txData.Data.Outputs) > 0 {
        script.Price = strconv.FormatUint(txData.Data.Outputs[0].Amount, 10)
    }
    if (len(txData.Data.Outputs) > 1 && index == 0) {
        script.To = txData.Data.Outputs[1].VerboseData.ScriptPublicKeyAddress
    } else {
        script.Price = "0"
        script.To = script.From
    }
}

////////////////////////////////
func (opMethodSend OpMethodSend) Validate(script *storage.DataScriptType, daaScore uint64, testnet bool) (bool) {
    if (!testnet && daaScore < 97539090) {
        return false
    }
    if (script.From == "" || script.To == "" || script.Utxo == "" || script.P != "KRC-20" || !ValidateTick(&script.Tick)) {
        return false
    }
    script.Amt = ""
    script.Max = ""
    script.Lim = ""
    script.Pre = ""
    script.Dec = ""
    script.Mod = ""
    return true
}

////////////////////////////////
func (opMethodSend OpMethodSend) PrepareStateKey(opScript *storage.DataScriptType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opScript.Tick] = nil
    stateMap.StateBalanceMap[opScript.From+"_"+opScript.Tick] = nil
    stateMap.StateBalanceMap[opScript.To+"_"+opScript.Tick] = nil
    dataUtxo := strings.Split(opScript.Utxo, "_")
    stateMap.StateMarketMap[opScript.Tick+"_"+opScript.From+"_"+dataUtxo[0]] = nil
}

////////////////////////////////
func (opMethodSend OpMethodSend) Do(index int, opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    opScript := opData.OpScript[index]
    ////////////////////////////////
    if stateMap.StateTokenMap[opScript.Tick] == nil {
        opData.OpAccept = -1
        opData.OpError = "tick not found"
        return nil
    }
    ////////////////////////////////
    dataUtxo := strings.Split(opScript.Utxo, "_")
    keyMarket := opScript.Tick +"_"+ opScript.From +"_"+ dataUtxo[0]
    keyBalanceFrom := opScript.From +"_"+ opScript.Tick
    keyBalanceTo := opScript.To +"_"+ opScript.Tick
    stMarket := stateMap.StateMarketMap[keyMarket]
    stBalanceFrom := stateMap.StateBalanceMap[keyBalanceFrom]
    stBalanceTo := stateMap.StateBalanceMap[keyBalanceTo]
    nTickAffc := int64(0)
    ////////////////////////////////
    if stMarket == nil {
        opData.OpAccept = -1
        opData.OpError = "order not found"
        return nil
    }
    if stBalanceFrom == nil {
        opData.OpAccept = -1
        opData.OpError = "order abnormal"
        return nil
    }
    opScript.Amt = stMarket.TAmt
    amtBig := new(big.Int)
    amtBig.SetString(opScript.Amt, 10)
    lockedBig := new(big.Int)
    lockedBig.SetString(stBalanceFrom.Locked, 10)
    if amtBig.Cmp(lockedBig) > 0 {
        opData.OpAccept = -1
        opData.OpError = "order abnormal"
        return nil
    }
    ////////////////////////////////
    opData.StBefore = AppendStLineBalance(opData.StBefore, keyBalanceFrom, stBalanceFrom, false)
    if keyBalanceFrom != keyBalanceTo {
        opData.StBefore = AppendStLineBalance(opData.StBefore, keyBalanceTo, stBalanceTo, false)
    }
    opData.StBefore = AppendStLineMarket(opData.StBefore, keyMarket, stMarket, false)
    ////////////////////////////////
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
    lockedBig = lockedBig.Sub(lockedBig, amtBig)
    stBalanceFrom.Locked = lockedBig.Text(10)
    balanceBig := new(big.Int)
    balanceBig.SetString(stBalanceTo.Balance, 10)
    balanceBig = balanceBig.Add(balanceBig, amtBig)
    stBalanceTo.Balance = balanceBig.Text(10)
    if (stBalanceFrom.Balance == "0" && stBalanceFrom.Locked == "0") {
        nTickAffc --
    }
    stateMap.StateMarketMap[keyMarket] = nil
    ////////////////////////////////
    opData.SsInfo.TickAffc = AppendSsInfoTickAffc(opData.SsInfo.TickAffc, opScript.Tick, nTickAffc)
    balanceBig.SetString(stBalanceFrom.Balance, 10)
    balanceBig = balanceBig.Add(balanceBig, lockedBig)
    opData.SsInfo.AddressAffc = AppendSsInfoAddressAffc(opData.SsInfo.AddressAffc, keyBalanceFrom, balanceBig.Text(10))
    if keyBalanceFrom != keyBalanceTo {
        balanceBig.SetString(stBalanceTo.Balance, 10)
        lockedBig.SetString(stBalanceTo.Locked, 10)
        balanceBig = balanceBig.Add(balanceBig, lockedBig)
        opData.SsInfo.AddressAffc = AppendSsInfoAddressAffc(opData.SsInfo.AddressAffc, keyBalanceTo, balanceBig.Text(10))
    }
    ////////////////////////////////
    opData.StAfter = AppendStLineBalance(opData.StAfter, keyBalanceFrom, stBalanceFrom, true)
    opData.StAfter = AppendStLineBalance(opData.StAfter, keyBalanceTo, stBalanceTo, true)
    opData.StAfter = AppendStLineMarket(opData.StAfter, keyMarket, nil, true)
    ////////////////////////////////
    if (stBalanceFrom.Balance == "0" && stBalanceFrom.Locked == "0") {
        stateMap.StateBalanceMap[keyBalanceFrom] = nil
    }
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
/*func (opMethodSend OpMethodSend) UnDo() (error) {
    // ...
    return nil
}*/

// ...

