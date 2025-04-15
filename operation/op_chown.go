
////////////////////////////////
package operation

import (
    "kasplex-executor/misc"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethodChown struct {}

////////////////////////////////
func init() {
    opName := "chown"
    P_Registered["KRC-20"] = true
    Op_Registered[opName] = true
    Method_Registered[opName] = new(OpMethodChown)
}

////////////////////////////////
func (opMethodChown OpMethodChown) FeeLeast(daaScore uint64) (uint64) {
    // if daaScore ...
    return 0
}

////////////////////////////////
func (opMethodChown OpMethodChown) ScriptCollectEx(index int, script *storage.DataScriptType, txData *storage.DataTransactionType, testnet bool) {}

////////////////////////////////
func (opMethodChown OpMethodChown) Validate(script *storage.DataScriptType, txId string, daaScore uint64, testnet bool) (bool) {
    if (!testnet && daaScore < 110165000) {
        return false
    }
    if ValidateTxId(&script.Ca) {
        script.Tick = script.Ca
    }
    if (script.From == "" || script.To == "" || script.P != "KRC-20" || !ValidateTickTxId(&script.Tick)) {
        return false
    }
    script.Amt = ""
    script.Max = ""
    script.Lim = ""
    script.Pre = ""
    script.Dec = ""
    script.Utxo = ""
    script.Price = ""
    script.Mod = ""
    script.Name = ""
    script.Ca = ""
    return true
}

////////////////////////////////
func (opMethodChown OpMethodChown) PrepareStateKey(opScript *storage.DataScriptType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opScript.Tick] = nil
}

////////////////////////////////
func (opMethodChown OpMethodChown) Do(index int, opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
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
    if (opScript.To == stateMap.StateTokenMap[opScript.Tick].To || !misc.VerifyAddr(opScript.To, testnet)) {
        opData.OpAccept = -1
        opData.OpError = "address invalid"
        return nil
    }
    ////////////////////////////////
    stToken := stateMap.StateTokenMap[opScript.Tick]
    opScript.Name = stToken.Name
    ////////////////////////////////
    opData.StBefore = nil
    opData.StBefore = AppendStLineToken(opData.StBefore, opScript.Tick, stToken, true, false)
    ////////////////////////////////
    stToken.To = opScript.To
    stToken.OpMod = opData.OpScore
    stToken.MtsMod = opData.MtsAdd
    ////////////////////////////////
    opData.StAfter = nil
    opData.StAfter = AppendStLineToken(opData.StAfter, opScript.Tick, stToken, true, true)
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
/*func (opMethodChown OpMethodChown) UnDo() (error) {
    // ...
    return nil
}*/

// ...
