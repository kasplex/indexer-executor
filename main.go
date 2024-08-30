
////////////////////////////////
package main

import (
    "os"
    "fmt"
    "log"
    "sync"
    "strings"
    "syscall"
    "context"
    "log/slog"
    "os/signal"
    "path/filepath"
    "kasplex-executor/config"
    "kasplex-executor/storage"
    "kasplex-executor/explorer"
)

////////////////////////////////
func main() {
    fmt.Println("KASPlex Executor v"+config.Version)
    
    // Set the correct working directory.
    arg0 := os.Args[0]
    if strings.Index(arg0, "go-build") < 0 {
        dir, err := filepath.Abs(filepath.Dir(arg0))
        if err != nil {
            log.Fatalln("main fatal:", err.Error())
        }
        os.Chdir(dir)
    }
    
    // Use the file lock for startup.
    fLock := "./.lockExecutor"
    lock, err := os.Create(fLock)
    if err != nil {
        log.Fatalln("main fatal:", err.Error())
    }
    defer os.Remove(fLock)
    defer lock.Close()
    err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
    if err != nil {
        log.Fatalln("main fatal:", err.Error())
    }
    defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

    // Load config.
    var cfg config.Config
    config.Load(&cfg)
    
    // Set the log level.
    logOpt := &slog.HandlerOptions{Level: slog.LevelError,}
    if cfg.Debug == 3 {
        logOpt = &slog.HandlerOptions{Level: slog.LevelDebug,}
    } else if cfg.Debug == 2 {
        logOpt = &slog.HandlerOptions{Level: slog.LevelInfo,}
    } else if cfg.Debug == 1 {
        logOpt = &slog.HandlerOptions{Level: slog.LevelWarn,}
    }
    logHandler := slog.NewTextHandler(os.Stdout, logOpt)
    slog.SetDefault(slog.New(logHandler))
    
    // Set exit signal.
    ctx, cancel := context.WithCancel(context.Background())
    wg := &sync.WaitGroup{}
    wg.Add(1)
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    down := false
    go func() {
        <-c
        slog.Info("main stopping ..")
        cancel()
        down = true
        wg.Done()
    }()
    
    // Init storage driver.
    storage.Init(cfg.Cassandra, cfg.Rocksdb)
    
    // Init explorer if api server up.
    if (!down) {
        explorer.Init(ctx, wg, cfg.Startup, cfg.Testnet)
        go explorer.Run()
    }
    
    // Waiting
    wg.Wait()
}
