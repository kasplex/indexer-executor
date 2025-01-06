
////////////////////////////////
package operation

import (
    "math/big"
    "kasplex-executor/misc"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethodMint struct {}

////////////////////////////////
func init() {
    opName := "mint"
    P_Registered["KRC-20"] = true
    Op_Registered[opName] = true
    Method_Registered[opName] = new(OpMethodMint)
}

////////////////////////////////
func (opMethodMint OpMethodMint) FeeLeast(daaScore uint64) (uint64) {
    // if daaScore ...
    return 100000000
}

////////////////////////////////
func (opMethodMint OpMethodMint) ScriptCollectEx(index int, script *storage.DataScriptType, txData *storage.DataTransactionType, testnet bool) {}

////////////////////////////////
func (opMethodMint OpMethodMint) Validate(script *storage.DataScriptType, txId string, daaScore uint64, testnet bool) (bool) {
    if (script.From == "" || script.P != "KRC-20" || !ValidateTick(&script.Tick)) {
        return false
    }
    script.Amt = ""
    if script.To == "" {
        script.To = script.From
    }
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
func (opMethodMint OpMethodMint) PrepareStateKey(opScript *storage.DataScriptType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opScript.Tick] = nil
    stateMap.StateBalanceMap[opScript.To+"_"+opScript.Tick] = nil
}

////////////////////////////////
func (opMethodMint OpMethodMint) Do(index int, opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    opScript := opData.OpScript[index]
    ////////////////////////////////
    if stateMap.StateTokenMap[opScript.Tick] == nil {
        opData.OpAccept = -1
        opData.OpError = "tick not found"
        return nil
    }
    if stateMap.StateTokenMap[opScript.Tick].Mod != "" {
        opData.OpAccept = -1
        opData.OpError = "mode invalid"
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
    if !misc.VerifyAddr(opScript.To, testnet) {
        opData.OpAccept = -1
        opData.OpError = "address invalid"
        return nil
    }
    ////////////////////////////////
    keyBalance := opScript.To +"_"+ opScript.Tick
    stToken := stateMap.StateTokenMap[opScript.Tick]
    stBalance := stateMap.StateBalanceMap[keyBalance]
    ////////////////////////////////
    amt := stToken.Lim
    maxBig := new(big.Int)
    maxBig.SetString(stToken.Max, 10)
    mintedBig := new(big.Int)
    mintedBig.SetString(stToken.Minted, 10)
    leftBig := maxBig.Sub(maxBig, mintedBig)
    limBig := new(big.Int)
    limBig.SetString("0", 10)
    if limBig.Cmp(leftBig) >= 0 {
        opData.OpAccept = -1
        opData.OpError = "mint finished"
        return nil
    }
    limBig.SetString(amt, 10)
    if limBig.Cmp(leftBig) > 0 {
        amt = leftBig.Text(10)
    }
    opScript.Amt = amt
    limBig.SetString(amt, 10)
    mintedBig = mintedBig.Add(mintedBig, limBig)
    minted := mintedBig.Text(10)
    ////////////////////////////////
    opData.StBefore = nil
    opData.StBefore = AppendStLineToken(opData.StBefore, opScript.Tick, stToken, false, false)
    opData.StBefore = AppendStLineBalance(opData.StBefore, keyBalance, stBalance, false)
    ////////////////////////////////
    stToken.Minted = minted
    stToken.OpMod = opData.OpScore
    stToken.MtsMod = opData.MtsAdd
    if stBalance == nil {
        stBalance = &storage.StateBalanceType{
            Address: opScript.To,
            Tick: opScript.Tick,
            Dec: stToken.Dec,
            Balance: "0",
            Locked: "0",
            OpMod: opData.OpScore,
        }
        stateMap.StateBalanceMap[keyBalance] = stBalance
        ////////////////////////////
        opData.SsInfo.TickAffc = AppendSsInfoTickAffc(opData.SsInfo.TickAffc, opScript.Tick, 1)
    } else {
        ////////////////////////////
        opData.SsInfo.TickAffc = AppendSsInfoTickAffc(opData.SsInfo.TickAffc, opScript.Tick, 0)
    }
    mintedBig.SetString(stBalance.Balance, 10)
    mintedBig = mintedBig.Add(mintedBig, limBig)
    stBalance.Balance = mintedBig.Text(10)
    stBalance.OpMod = opData.OpScore
    ////////////////////////////////
    lockedBig := new(big.Int)
    lockedBig.SetString(stBalance.Locked, 10)
    mintedBig = mintedBig.Add(mintedBig, lockedBig)
    balanceTotal := mintedBig.Text(10)
    opData.SsInfo.AddressAffc = AppendSsInfoAddressAffc(opData.SsInfo.AddressAffc, keyBalance, balanceTotal)
    ////////////////////////////////
    opData.StAfter = nil
    opData.StAfter = AppendStLineToken(opData.StAfter, opScript.Tick, stToken, false, true)
    opData.StAfter = AppendStLineBalance(opData.StAfter, keyBalance, stBalance, true)
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
/*func (opMethodMint OpMethodMint) UnDo() (error) {
    // ...
    return nil
}*/

// ...
