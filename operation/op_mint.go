
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
func (opMethodMint OpMethodMint) Validate(script *storage.DataScriptType, testnet bool) (bool) {
    if (script.From == "" || script.P != "KRC-20" || !ValidateTick(&script.Tick)) {
        return false
    }
    script.Amt = ""
    if script.To == "" {
        script.To = script.From
    }
    return true
}

////////////////////////////////
func (opMethodMint OpMethodMint) PrepareStateKey(opData *storage.DataOperationType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opData.OpScript.Tick] = nil
    stateMap.StateBalanceMap[opData.OpScript.To+"_"+opData.OpScript.Tick] = nil
}

////////////////////////////////
func (opMethodMint OpMethodMint) Do(opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    ////////////////////////////////
    if stateMap.StateTokenMap[opData.OpScript.Tick] == nil {
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
    if !misc.VerifyAddr(opData.OpScript.To, testnet) {
        opData.OpAccept = -1
        opData.OpError = "address invalid"
        return nil
    }
    ////////////////////////////////
    keyBalance := opData.OpScript.To +"_"+ opData.OpScript.Tick
    stToken := stateMap.StateTokenMap[opData.OpScript.Tick]
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
    opData.OpScript.Amt = amt
    limBig.SetString(amt, 10)
    mintedBig = mintedBig.Add(mintedBig, limBig)
    minted := mintedBig.Text(10)
    ////////////////////////////////
    opData.StBefore = nil
    stLine := MakeStLineToken(opData.OpScript.Tick, stToken, false)
    opData.StBefore = append(opData.StBefore, stLine)
    stLine = MakeStLineBalance(keyBalance, stBalance)
    opData.StBefore = append(opData.StBefore, stLine)
    ////////////////////////////////
    stToken.Minted = minted
    stToken.OpMod = opData.OpScore
    stToken.MtsMod = opData.MtsAdd
    if stBalance == nil {
        stBalance = &storage.StateBalanceType{
            Address: opData.OpScript.To,
            Tick: opData.OpScript.Tick,
            Dec: stToken.Dec,
            Balance: "0",
            Locked: "0",
            OpMod: opData.OpScore,
        }
        stateMap.StateBalanceMap[keyBalance] = stBalance
        ////////////////////////////
        opData.SsInfo.TickAffc = append(opData.SsInfo.TickAffc, opData.OpScript.Tick+"=1")
    } else {
        ////////////////////////////
        opData.SsInfo.TickAffc = append(opData.SsInfo.TickAffc, opData.OpScript.Tick+"=0")
    }
    mintedBig.SetString(stBalance.Balance, 10)
    mintedBig = mintedBig.Add(mintedBig, limBig)
    stBalance.Balance = mintedBig.Text(10)
    stBalance.OpMod = opData.OpScore
    ////////////////////////////////
    opData.SsInfo.AddressAffc = append(opData.SsInfo.AddressAffc, opData.OpScript.To+"_"+opData.OpScript.Tick+"="+stBalance.Balance)
    ////////////////////////////////
    opData.StAfter = nil
    stLine = MakeStLineToken(opData.OpScript.Tick, stToken, false)
    opData.StAfter = append(opData.StAfter, stLine)
    stLine = MakeStLineBalance(keyBalance, stBalance)
    opData.StAfter = append(opData.StAfter, stLine)
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
func (opMethodMint OpMethodMint) UnDo() (error) {
    // ...
    return nil
}

// ...
