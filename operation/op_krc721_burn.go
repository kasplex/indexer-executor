////////////////////////////////
package operation

import (
    "math/big"
    "kasplex-executor/misc"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethodBurnKRC721 struct {}

////////////////////////////////
func init() {
    opName := "burnKRC721"
    P_Registered["KRC-721"] = true
    Op_Registered[opName] = true
    Method_Registered[opName] = new(OpMethodBurnKRC721)
}

////////////////////////////////
func (opMethodBurnKRC721 OpMethodBurnKRC721) FeeLeast(daaScore uint64) (uint64) {
    return 0
}

////////////////////////////////
func (opMethodBurnKRC721 OpMethodBurnKRC721) ScriptCollectEx(index int, script *storage.DataScriptType, txData *storage.DataTransactionType, testnet bool) {}

////////////////////////////////
func (opMethodBurnKRC721 OpMethodBurnKRC721) Validate(script *storage.DataScriptType, daaScore uint64, testnet bool) (bool) {
    if (script.From == "" || script.P != "KRC-721" || !ValidateTick(&script.Tick) || !ValidateAmount(&script.Amt)) {
        return false
    }
    script.To = ""
    script.Max = ""
    script.Lim = ""
    script.Pre = ""
    script.Dec = ""
    script.Utxo = ""
    script.Price = ""
    return true
}

////////////////////////////////
func (opMethodBurnKRC721 OpMethodBurnKRC721) PrepareStateKey(opScript *storage.DataScriptType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opScript.Tick] = nil
    stateMap.StateBalanceMap[opScript.From+"_"+opScript.Tick] = nil
}

////////////////////////////////
func (opMethodBurnKRC721 OpMethodBurnKRC721) Do(index int, opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    opScript := opData.OpScript[index]
    ////////////////////////////////
    if stateMap.StateTokenMap[opScript.Tick] == nil {
        opData.OpAccept = -1
        opData.OpError = "tick not found"
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
    ////////////////////////////////
    keyBalance := opScript.From +"_"+ opScript.Tick
    stToken := stateMap.StateTokenMap[opScript.Tick]
    stBalance := stateMap.StateBalanceMap[keyBalance]
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
    }
    ////////////////////////////////
    opData.StBefore = nil
    opData.StBefore = AppendStLineToken(opData.StBefore, opScript.Tick, stToken, false, false)
    opData.StBefore = AppendStLineBalance(opData.StBefore, keyBalance, stBalance, false)
    ////////////////////////////////
    balanceBig = balanceBig.Sub(balanceBig, amtBig)
    stBalance.Balance = balanceBig.Text(10)
    stBalance.OpMod = opData.OpScore
    mintedBig := new(big.Int)
    mintedBig.SetString(stToken.Minted, 10)
    mintedBig = mintedBig.Sub(mintedBig, amtBig)
    stToken.Minted = mintedBig.Text(10)
    stToken.OpMod = opData.OpScore
    stToken.MtsMod = opData.MtsAdd
    ////////////////////////////////
    lockedBig := new(big.Int)
    lockedBig.SetString(stBalance.Locked, 10)
    balanceBig = balanceBig.Add(balanceBig, lockedBig)
    balanceTotal := balanceBig.Text(10)
    opData.SsInfo.AddressAffc = AppendSsInfoAddressAffc(opData.SsInfo.AddressAffc, opScript.From+"_"+opScript.Tick, balanceTotal)
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
/*func (opMethodBurnKRC721 OpMethodBurnKRC721) UnDo() (error) {
    // ...
    return nil
}*/
