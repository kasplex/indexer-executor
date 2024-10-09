# Kasplex indexer-executor installation

**Note:** All commands in this document were executed on Linux Ubuntu.

## 1. Install dependencies that might be needed.
```shell
sudo apt install -y libsnappy-dev libgflags-dev unzip make build-essential
```
***

## 2. Install Rocksdb

2.1 download rocksdb of version: 6.15.5
```shell
cd /var
sudo wget https://github.com/facebook/rocksdb/archive/refs/tags/v6.15.5.zip -O ./rocksdb.6155.zip
```

2.2 unzip the rocksdb
```shell
sudo unzip -o -d /var ./rocksdb.6155.zip
```

2.3 make the rocksdb
```shell
cd /var/rocksdb-6.15.5
sudo PORTABLE=1 make shared_lib
sudo INSTALL_PATH=/usr/lib make install-shared
```

***

## 3. Install Golang
3.1 Here we recommend that you install the latest version of Golang
```shell
sudo snap install go --classic
```

3.2 If you need to install a specific version of Golang, you can use the following command:
Replace <version> with the desired version number, for example:
```shell
sudo snap install go --channel=<version>
```

***

## 4. Project Initialization

4.1 Download code from github
```shell
git clone https://github.com/kasplex/indexer-executor.git
```

4.2 build the program of executor
```shell
cd indexer-executor
go get -u && go mod tidy
go build -o kpexecutor main.go
```

***

## 5. Run Project

5.1 Create a configuration file named config.json. Use config.sample.mainnet.json as a reference. The configuration details are as follows:
```shell
{
    "startup": {                              // executor job config
        "hysteresis": 3,                      // hysteresis indicates the number of blocks delayed from the latest block, with a suggested value of 3. A setting of 0 means no lag, providing the highest real-time performance but is more prone to handling higher numbers of rollbacks.
        "start": "",                          // the hash of the sync start block, only when testnet=true.
        "daaScoreRange": [],                  // just The range for executing daascore, only when testnet=true.
        "tickReserved": []                    // The reserved tick address list, only when testnet=true.
    },
    "cassandra": {                            // cassandra config
        "host": "",                           // connection host           
        "port": 9142,                         // connection port
        "user": "",                           // connection user
        "pass": "",                           // connection password
        "crt": "sf-class2-root.crt",          // file path of db access key 
        "space": "kpdb01mainnet"              // connection name
    },
    "rocksdb": {                              // This part is the local database parameters.
        "path": "./data"                      // db path
    },
    "testnet": false,                         //true: mainnet  false: testnet
    "debug": 2                                //log level:  1:Warn  2: Info  3:Debug 
}
```

5.2 Run manually.
```shell
sudo ./kpexecutor
```

***

> **Note:** The following section on configuring system services is optional.

## 6. Configure system services (optional).

6.1 Create the /etc/systemd/system/kasplex-executor.service file, and set the content as below, remember to replace the correct project path inside.
```shell
[Unit]
Description=Kasplex Executor Service
After=network.target
[Service]
WorkingDirectory=/{rootpath}
ExecStart=/{rootpath}/kpexecutor
Restart=always
RestartSec=10s
User=root
Group=root
[Install]
WantedBy=multi-user.target
```

6.2 Update service configuration.
```shell
sudo systemctl daemon-reload
sudo systemctl enable kasplex-executor.service
```

6.3 Start service.
```shell
sudo systemctl start kasplex-executor.service
```

