////////////////////////////////
package config

import (
    "os"
    "log"
    "encoding/json"
)

////////////////////////////////
type StartupConfig struct {
    Hysteresis int `json:"hysteresis"`
    DaaScoreRange [][2]uint64 `json:"daaScoreRange"`
    TickReserved []string `json:"tickReserved"`
}
type CassaConfig struct {
    Host string `json:"host"`
    Port int `json:"port"`
    User string `json:"user"`
    Pass string `json:"pass"`
    Crt string `json:"crt"`
    Space string `json:"space"`
}
type RocksConfig struct {
    Path string `json:"path"`
}
type Config struct {
    Startup StartupConfig `json:"startup"`
    Cassandra CassaConfig `json:"cassandra"`
    Rocksdb RocksConfig `json:"rocksdb"`
    Debug int `json:"debug"`
    Testnet bool `json:"testnet"`
    KRC721Enabled bool `json:"krc721Enabled"`
    KRC721ContractAddress string `json:"krc721ContractAddress"`
}

////////////////////////////////
const Version = "2.02.1130"

////////////////////////////////
func Load(cfg *Config) {

    // File "config.json" should be in the same directory.
    
    dir, _ := os.Getwd()
    fp, err := os.Open(dir + "/config.json")
    if err != nil {
        log.Fatalln("config.Load fatal:", err.Error())
    }
    defer fp.Close()
    jParser := json.NewDecoder(fp)
    err = jParser.Decode(&cfg)
    if err != nil {
        log.Fatalln("config.Load fatal:", err.Error())
    }
}
