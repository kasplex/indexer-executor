
////////////////////////////////
package storage

import (
    "log"
    "strings"
    "log/slog"
    "github.com/gocql/gocql"
    "github.com/tecbot/gorocksdb"
    "kasplex-executor/config"
)

////////////////////////////////
type runtimeType struct {
    cassa *gocql.ClusterConfig
    sessionCassa *gocql.Session
    rocksTx *gorocksdb.TransactionDB
    rOptRocks *gorocksdb.ReadOptions
    wOptRocks *gorocksdb.WriteOptions
    txOptRocks *gorocksdb.TransactionOptions
    cfgCassa config.CassaConfig
    cfgRocks config.RocksConfig
    // ...
}
var sRuntime runtimeType

////////////////////////////////
func Init(cfgCassa config.CassaConfig, cfgRocks config.RocksConfig) {
    sRuntime.cfgCassa = cfgCassa
    sRuntime.cfgRocks = cfgRocks
    slog.Info("storage.Init start.")
    
    // Use cassandra driver.
    var err error
    sRuntime.cassa = gocql.NewCluster(sRuntime.cfgCassa.Host)
    sRuntime.cassa.Port = sRuntime.cfgCassa.Port
    sRuntime.cassa.Authenticator = gocql.PasswordAuthenticator{
        Username: sRuntime.cfgCassa.User,
        Password: sRuntime.cfgCassa.Pass,
    }
    if sRuntime.cfgCassa.Crt != "" {
        sRuntime.cassa.SslOpts = &gocql.SslOptions{
            CaPath: sRuntime.cfgCassa.Crt,
            EnableHostVerification: false,
        }
    }
    sRuntime.cassa.Consistency = gocql.LocalQuorum
    sRuntime.cassa.DisableInitialHostLookup = false
    sRuntime.cassa.NumConns = nBatchMaxCassa
    sRuntime.cassa.Keyspace = sRuntime.cfgCassa.Space
    sRuntime.sessionCassa, err = sRuntime.cassa.CreateSession()
    if err != nil {
        log.Fatalln("storage.Init fatal: ", err.Error())
    }
    
    // Init database if new installation.
    for _, cqln := range cqlnInitTable {
        err = sRuntime.sessionCassa.Query(cqln).Exec()
        if err != nil {
            if strings.HasSuffix(err.Error(), "conflicts with an existing column") {
                continue
            }
            log.Fatalln("storage.Init fatal:", err.Error())
        }
    }
    
    // Use rocksdb driver.
    sRuntime.rOptRocks = gorocksdb.NewDefaultReadOptions()
    sRuntime.wOptRocks = gorocksdb.NewDefaultWriteOptions()
    sRuntime.txOptRocks = gorocksdb.NewDefaultTransactionOptions()
    optRocks := gorocksdb.NewDefaultOptions()
    optRocks.SetUseFsync(true)
    optRocks.SetCreateIfMissing(true)
    optRocks.SetCreateIfMissingColumnFamilies(true)
    optRocks.SetWriteBufferSize(64 * 1024 * 1024)
    optRocks.SetMaxWriteBufferNumber(3)
    optRocks.SetMaxBackgroundCompactions(4)
    optBbtRocks := gorocksdb.NewDefaultBlockBasedTableOptions()
    optBbtRocks.SetBlockSize(8 * 1024)
    optBbtRocks.SetBlockCache(gorocksdb.NewLRUCache(64 * 1024 * 1024))
    optBbtRocks.SetFilterPolicy(gorocksdb.NewBloomFilter(10))
    optRocks.SetBlockBasedTableFactory(optBbtRocks)
    txOptRocks := gorocksdb.NewDefaultTransactionDBOptions()
    txOptRocks.SetTransactionLockTimeout(10)
    sRuntime.rocksTx, err = gorocksdb.OpenTransactionDb(optRocks, txOptRocks, sRuntime.cfgRocks.Path)
    if err != nil {
        log.Fatalln("storage.Init fatal: ", err.Error())
    }
        
    slog.Info("storage ready.")
}

