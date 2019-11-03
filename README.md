# Qitmeer Miner

    The Miner of Qitmeer
[![Build Status](https://travis-ci.com/Qitmeer/qitmeer-miner.svg?token=n9AoZUDqAJmhesf4MYUd&branch=master)](https://travis-ci.com/Qitmeer/qitmeer-miner)
## Enviroment

```bash
$ go version >= 1.12
$ cargo >= 1.36.0 (c4fcfb725 2019-05-15)
```
    
## Compile

```bash
$ git clone git@github.com:Qitmeer/qitmeer-miner.git
$ cd lib/cuckoo
$ cargo build --release
```

* Ubuntu ENV
```bash
$ sudo apt-get install beignet-dev nvidia-cuda-dev nvidia-cuda-toolkit
$ sudo copy lib/cuckoo/target/release/libcuckoo.so /usr/lib/
$ go build 
```
        
* Centos ENV
```bash
$ sudo yum install opencl-headers
$ sudo yum install ocl-icd
$ sudo ln -s /usr/lib64/libOpenCL.so.1 /usr/lib/libOpenCL.so
$ sudo copy lib/cuckoo/target/release/libcuckoo.so /usr/lib/
$ go build
```   

* MAC

```bash
go build
```

* Windows ENV
##### Prequisite: 
1. Install **Build Tools for Visual Studio**:  
https://visualstudio.microsoft.com/thank-you-downloading-visual-studio/?sku=BuildTools&rel=16

2. build
```bash
$ copy lib/cuckoo/target/release/cuckoo.dll to C:/Windows
$ go build 
```

* Other method
```bash
$ make
```

###### Any questions see [docs](https://qitmeer.github.io/docs/en/reference/qitmeer-miner/)
###### You can also see mining [tutorial](https://www.qitmeertalk.org/t/qitmeer-2019-11-02/906)

## Run
```bash
$ cp qitmeer.conf.example qitmeer.conf
```
- 1.run with the config file `qitmeer.conf`
- 2.run with command
```bash
$ ./qitmeer-miner -h
Usage:
    qitmeer-miner [OPTIONS]

Debug Command:
  -l, --listdevices    List number of devices.

The Config File Options:
  -C, --configfile=    Path to configuration file
      --minerlog=      Write miner log file

The Necessary Config Options:
  -P, --pow=           blake2bd|cuckaroo|cuckatoo (blake2bd)
  -S, --symbol=        Symbol (PMEER)
  -N, --network=       network privnet|testnet|mainnet (mainnet)

The Solo Config Option:
  -M, --mineraddress=  Miner Address
  -s, --rpcserver=     RPC server to connect to (127.0.0.1)
  -u, --rpcuser=       RPC username
  -p, --rpcpass=       RPC password
      --randstr=       Rand String,Your Unique Marking. (Come from Qitmeer!)
      --notls          Do not verify tls certificates (true)
      --rpccert=       RPC server certificate chain for validation

The pool Config Option:
  -o, --pool=          Pool to connect to (e.g.stratum+tcp://pool:port)
  -m, --pooluser=      Pool username
  -n, --poolpass=      Pool password

The Optional Config Option:
      --cpuminer       CPUMiner (false)
      --proxy=         Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)
      --proxyuser=     Username for proxy server
      --proxypass=     Password for proxy server
      --trimmerTimes=  the cuckaroo trimmer times (40)
      --intensity=     Intensities (the work size is 2^intensity) per device. Single global value or a comma
                       separated list. (24)
      --worksize=      The explicitly declared sizes of the work to do per device (overrides intensity). Single
                       global value or a comma separated list. (256)
      --timeout=       rpc timeout. (60)
      --use_devices=   all gpu devices,you can use ./qitmeer-miner -l to see. examples:0,1 use the #0 device
                       and #1 device
      --max_tx_count=  max pack tx count (1000)
      --max_sig_count= max sign tx count (5000)
      --log_level=     info|debug|error|warn|trace (debug)
      --stats_server=  stats web server (127.0.0.1:1235)
      --edge_bits=     edge bits (24)
      --local_size=    local size (4096)
      --group_size=    work group size (256)

Help Options:
  -h, --help           Show this help message
```
## Stats Web Server
- add param `stats_server=127.0.0.1:1235` in qitmeer.conf
- brower explorer http://127.0.0.1:1235    
![stats](public/img/miner1.png)  
![stats](public/img/miner2.png)  
![stats](public/img/miner3.png)  
