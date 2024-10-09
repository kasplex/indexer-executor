# Kasplex indexer-executor installation

**Note:** All commands in this document were executed on Linux Ubuntu.

## 1. Install dependencies that might be needed.
```shell
sudo apt install libsnappy-dev libgflags-dev unzip make build-essential
```

## 2. Install Rocksdb

2.1 download rocksdb of version: 6.15.5
```shell
cd /var
wget https://github.com/facebook/rocksdb/archive/refs/tags/v6.15.5.zip -O ./rocksdb.6155.zip
```

2.2 unzip the rocksdb
```shell
sudo unzip -o -d /var ./rocksdb.6155.zip
cd /var/rocksdb-6.15.5
```

2.3 make the rocksdb
```shell
sudo PORTABLE=1 make shared_lib
sudo INSTALL_PATH=/usr/lib make install-shared
```
 
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

## 4. Project Initialization

4.1 Download code from github
```shell
git clone https://github.com/******.git
```

4.2 build the program of executor
```shell
cd indexer-executor
go get -u && go mod tidy
go build -o kpexecutor main.go
```