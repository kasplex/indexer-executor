
////////////////////////////////
package storage

import (
    "kasplex-executor/protowire"
)

////////////////////////////////
type DataVspcType struct {
    DaaScore uint64 `json:"daaScore"`
    Hash string `json:"hash"`
    TxIdList []string `json:"-"`
}

////////////////////////////////
type DataTransactionType struct {
    TxId string
    DaaScore uint64
    BlockAccept string
    Data *protowire.RpcTransaction
}

////////////////////////////////
type DataScriptType struct {
    P string `json:"p"`
    Op string `json:"op"`
    From string `json:"from,omitempty"`
    To string `json:"to,omitempty"`
    Tick string `json:"tick,omitempty"`
    Max string `json:"max,omitempty"`
    Lim string `json:"lim,omitempty"`
    Pre string `json:"pre,omitempty"`
    Dec string `json:"dec,omitempty"`
    Amt string `json:"amt,omitempty"`
    Utxo string `json:"utxo,omitempty"`
    Price string `json:"price,omitempty"`
    Mod string `json:"mod,omitempty"`
    Name string `json:"name,omitempty"`
    // ...
}

////////////////////////////////
type DataOpStateType struct {
    BlockAccept string `json:"blockaccept,omitempty"`
    Fee uint64 `json:"fee,omitempty"`
    FeeLeast uint64 `json:"feeleast,omitempty"`
    MtsAdd int64 `json:"mtsadd,omitempty"`
    OpScore uint64 `json:"opscore,omitempty"`
    OpAccept int8 `json:"opaccept,omitempty"`
    OpError string `json:"operror,omitempty"`
    Checkpoint string `json:"checkpoint,omitempty"`
}

////////////////////////////////
type DataStatsType struct {
    TickAffc []string
    AddressAffc []string
    // XxxAffc ...
}

////////////////////////////////
type DataOperationType struct {
    TxId string
    DaaScore uint64
    BlockAccept string
    Fee uint64
    FeeLeast uint64
    MtsAdd int64
    OpScore uint64
    OpAccept int8
    OpError string
    OpScript []*DataScriptType
    ScriptSig string
    StBefore []string
    StAfter []string
    Checkpoint string
    SsInfo *DataStatsType
}

////////////////////////////////
type StateTokenMetaType struct {
    Max string `json:"max,omitempty"`
    Lim string `json:"lim,omitempty"`
    Pre string `json:"pre,omitempty"`
    Dec int `json:"dec,omitempty"`
    Mod string `json:"mod,omitempty"`
    From string `json:"from,omitempty"`
    To string `json:"to,omitempty"`
    Name string `json:"name,omitempty"`
    TxId string `json:"txid,omitempty"`
    OpAdd uint64 `json:"opadd,omitempty"`
    MtsAdd int64 `json:"mtsadd,omitempty"`
}

////////////////////////////////
type StateTokenType struct {
    Tick string `json:"tick,omitempty"`
    Max string `json:"max,omitempty"`
    Lim string `json:"lim,omitempty"`
    Pre string `json:"pre,omitempty"`
    Dec int `json:"dec,omitempty"`
    Mod string `json:"mod,omitempty"`
    From string `json:"from,omitempty"`
    To string `json:"to,omitempty"`
    Minted string `json:"minted,omitempty"`
    Burned string `json:"burned,omitempty"`
    Name string `json:"name,omitempty"`
    TxId string `json:"txid,omitempty"`
    OpAdd uint64 `json:"opadd,omitempty"`
    OpMod uint64 `json:"opmod,omitempty"`
    MtsAdd int64 `json:"mtsadd,omitempty"`
    MtsMod int64 `json:"mtsmod,omitempty"`
}

////////////////////////////////
type StateBalanceType struct {
    Address string `json:"address,omitempty"`
    Tick string `json:"tick,omitempty"`
    Dec int `json:"dec,omitempty"`
    Balance string `json:"balance,omitempty"`
    Locked string `json:"locked,omitempty"`
    OpMod uint64 `json:"opmod,omitempty"`
}

////////////////////////////////
type StateMarketType struct {
    Tick string `json:"tick,omitempty"`
    TAddr string `json:"taddr,omitempty"`
    UTxId string `json:"utxid,omitempty"`
    UAddr string `json:"uaddr,omitempty"`
    UAmt string `json:"uamt,omitempty"`
    UScript string `json:"uscript,omitempty"`
    TAmt string `json:"tamt,omitempty"`
    OpAdd uint64 `json:"opadd,omitempty"`
}

////////////////////////////////
type StateBlacklistType struct {
    Tick string `json:"tick,omitempty"`
    Address string `json:"address,omitempty"`
    OpAdd uint64 `json:"opadd,omitempty"`
}

////////////////////////////////
// type StateXxx ...

////////////////////////////////
type DataStateMapType struct {
    StateTokenMap map[string]*StateTokenType `json:"statetokenmap,omitempty"`
    StateBalanceMap map[string]*StateBalanceType `json:"statebalancemap,omitempty"`
    StateMarketMap map[string]*StateMarketType `json:"statemarketmap,omitempty"`
    StateBlacklistMap map[string]*StateBlacklistType `json:"stateblacklistmap,omitempty"`
    // StateXxx ...
}

////////////////////////////////
type DataRollbackType struct {
    DaaScoreStart uint64 `json:"daascorestart"`
    DaaScoreEnd uint64 `json:"daascoreend"`
    CheckpointBefore string `json:"checkpointbefore"`
    CheckpointAfter string `json:"checkpointafter"`
    OpScoreLast uint64 `json:"opscorelast"`
    StateMapBefore DataStateMapType `json:"statemapbefore"`
    OpScoreList []uint64 `json:"opscorelist"`
    TxIdList []string `json:"txidlist"`
}

////////////////////////////////
type DataInputType struct {
    Hash string
    Index uint
    Amount uint64
}

////////////////////////////////
type DataFeeType struct {
    Txid string
    InputList []DataInputType
    AmountOut uint64
    Fee uint64
}

// ...
