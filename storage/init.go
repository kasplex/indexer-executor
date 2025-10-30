
////////////////////////////////
package storage

//#cgo CFLAGS: -I${SRCDIR}/rocksdb-6.15.5/include
//#cgo LDFLAGS: -L${SRCDIR}/rocksdb-6.15.5 -lrocksdb -lstdc++ -lm -lz -lsnappy -lzstd -llz4 -lbz2 -static
import "C"
import (
    "log"
    "time"
    "strings"
    "log/slog"
    "math/rand"
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
    rand.Seed(time.Now().UnixNano())
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
    sRuntime.cassa.NumConns = numConns
    sRuntime.cassa.Keyspace = sRuntime.cfgCassa.Space
    sRuntime.sessionCassa, err = sRuntime.cassa.CreateSession()
    if err != nil {
        log.Fatalln("storage.Init fatal: ", err.Error())
    }
    
    // Init database if new installation.
    for _, cqln := range cqlnInitTable {
        err = sRuntime.sessionCassa.Query(cqln).Exec()
        if err != nil {
            msg := err.Error()
            if strings.HasSuffix(msg, "conflicts with an existing column") || strings.HasSuffix(msg, "already exists") {
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
    optRocks.SetWriteBufferSize(256 * 1024 * 1024)
    optRocks.SetMaxWriteBufferNumber(4)
    optRocks.SetMaxBackgroundCompactions(4)
    optBbtRocks := gorocksdb.NewDefaultBlockBasedTableOptions()
    optBbtRocks.SetBlockSize(8 * 1024)
    optBbtRocks.SetBlockCache(gorocksdb.NewLRUCache(1024 * 1024 * 1024))
    optBbtRocks.SetFilterPolicy(gorocksdb.NewBloomFilter(10))
    optBbtRocks.SetCacheIndexAndFilterBlocks(true)
    optRocks.SetBlockBasedTableFactory(optBbtRocks)
    txOptRocks := gorocksdb.NewDefaultTransactionDBOptions()
    txOptRocks.SetTransactionLockTimeout(10)
    sRuntime.rocksTx, err = gorocksdb.OpenTransactionDb(optRocks, txOptRocks, sRuntime.cfgRocks.Path)
    if err != nil {
        log.Fatalln("storage.Init fatal: ", err.Error())
    }
        
    slog.Info("storage ready.")
}

