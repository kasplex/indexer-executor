
////////////////////////////////
package operation

import (
    "kasplex-executor/misc"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethodBlacklist struct {}

////////////////////////////////
func init() {
    opName := "blacklist"
    P_Registered["KRC-20"] = true
    Op_Registered[opName] = true
    Method_Registered[opName] = new(OpMethodBlacklist)
}

////////////////////////////////
func (opMethodBlacklist OpMethodBlacklist) FeeLeast(daaScore uint64) (uint64) {
    // if daaScore ...
    return 0
}

////////////////////////////////
func (opMethodBlacklist OpMethodBlacklist) ScriptCollectEx(index int, script *storage.DataScriptType, txData *storage.DataTransactionType, testnet bool) {}

////////////////////////////////
func (opMethodBlacklist OpMethodBlacklist) Validate(script *storage.DataScriptType, txId string, daaScore uint64, testnet bool) (bool) {
    if (!testnet && daaScore < 9999999999) {  // undetermined for mainnet
        return false
    }
    if (script.From == "" || script.To == "" || script.P != "KRC-20" || !ValidateTxId(&script.Ca)) {
        return false
    }
    if (script.Mod != "add" && script.Mod != "remove") {
        return false
    }
    script.Tick = script.Ca
    script.Amt = ""
    script.Max = ""
    script.Lim = ""
    script.Pre = ""
    script.Dec = ""
    script.Utxo = ""
    script.Price = ""
    script.Name = ""
    script.Ca = ""
    return true
}

////////////////////////////////
func (opMethodBlacklist OpMethodBlacklist) PrepareStateKey(opScript *storage.DataScriptType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opScript.Tick] = nil
    stateMap.StateBlacklistMap[opScript.Tick+"_"+opScript.To] = nil
}

////////////////////////////////
func (opMethodBlacklist OpMethodBlacklist) Do(index int, opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    opScript := opData.OpScript[index]
    ////////////////////////////////
    if stateMap.StateTokenMap[opScript.Tick] == nil {
        opData.OpAccept = -1
        opData.OpError = "tick not found"
        return nil
    }
    if stateMap.StateTokenMap[opScript.Tick].Mod != "issue" {
        opData.OpAccept = -1
        opData.OpError = "mode invalid"
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
    keyBlacklist := opScript.Tick +"_"+ opScript.To
    opScript.Name = stateMap.StateTokenMap[opScript.Tick].Name
    ////////////////////////////////
    if (opScript.Mod == "add" && stateMap.StateBlacklistMap[keyBlacklist] != nil) {
        opData.OpAccept = -1
        opData.OpError = "no affected"
        return nil
    }
    if (opScript.Mod == "remove" && stateMap.StateBlacklistMap[keyBlacklist] == nil) {
        opData.OpAccept = -1
        opData.OpError = "no affected"
        return nil
    }
    ////////////////////////////////
    opData.StBefore = nil
    opData.StBefore = AppendStLineBlacklist(opData.StBefore, keyBlacklist, stateMap.StateBlacklistMap[keyBlacklist], false)
    ////////////////////////////////
    if opScript.Mod == "add" {
        stateMap.StateBlacklistMap[keyBlacklist] = &storage.StateBlacklistType{
            Tick: opScript.Tick,
            Address: opScript.To,
            OpAdd: opData.OpScore,
        }
    } else if opScript.Mod == "remove" {
        stateMap.StateBlacklistMap[keyBlacklist] = nil
    }
    ////////////////////////////////
    opData.StAfter = nil
    opData.StAfter = AppendStLineBlacklist(opData.StAfter, keyBlacklist, stateMap.StateBlacklistMap[keyBlacklist], true)
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
/*func (opMethodBlacklist OpMethodBlacklist) UnDo() (error) {
    // ...
    return nil
}*/

// ...
