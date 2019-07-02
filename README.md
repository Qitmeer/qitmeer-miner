# HLC Miner

    The miner of Halalchain

## Enviroment

```bash
$ go version >= 1.12
$ go build
```
    
    
## Compile

```bash
$ git clone (this repo)
```

* Ubuntu ENV
```bash
$ sudo apt-get install beignet-dev nvidia-cuda-dev nvidia-cuda-toolkit
$ go build 
```
        
* Centos ENV
```bash
$ sudo yum install opencl-headers
$ sudo yum install ocl-icd
$ sudo ln -s /usr/lib64/libOpenCL.so.1 /usr/lib/libOpenCL.so
$ go build
```
        

* MAC

```bash
go build
```
    
* Windows ENV

  - install the opencl driver
```bash
$ go build 
```
        
    
## Run
```bash
$ cp halalchainminer.conf.example halalchainminer.conf
```
- 1.run with the config file `halalchainminer.conf`
    
```bash
# the miner config file
    
#node is dag
dag=true
    
#coin
symbol=HLC

#pow type  cuckaroo|blake2bd|cuckatoo  cuckatoo TODO
pow=blake2bd
    

#trimmerCount use with pow cuckaroo
trimmerCount=60

#not tls
notls=true

#rpccert the path of the node cert
#rpccert=CA.cert

#miner address
mineraddress=RmN4SADy42FKmN8ARKieX9iHh9icptdgYNn
    
#network simnet | testnet | mainnet
testnet=true
    
# Intensities (the work size is 2^intensity) up to device
intensity=24
    
# The explicitly declared sizes of the work to do up to device (overrides intensity)
worksize=256
    
########################## solo config ####################
# node rpc server
rpcserver=127.0.0.1:1234
    
# node rpc user
rpcuser=test
    
# node rpc password
rpcpass=test
    
########################## pool config ,if use this , it will use pool mining ########################
    
#pool=stratum+tcp://127.0.0.1:3177
#pooluser=RmN4SADy42FKmN8ARKieX9iHh9icptdgYNn
#poolpass=
    

```
    
```bash
$ ./hlc-miner
```
- 2.solo run with command line

```bash
$ ./hlc-miner -s 127.0.0.1:1234 -u test -P test --symbol HLC --notls -i 24 -W 256 --mineraddress RmN4SADy42FKmN8ARKieX9iHh9icptdgYNn --pow blake2bd
```
- 3.pool run command line

```bash
$ ./hlc-miner -o stratum+tcp://127.0.0.1:3177 -m RmN4SADy42FKmN8ARKieX9iHh9icptdgYNn --symbol HLC --notls -i 24 -W 256 --pow blake2bd
``` 

## Param Description 
          
- `--dag` the node is dag node
- `-s` the node rpc listen address
- `-u` the node rpc username
- `-P` the node rpc password
- `--symbol` now just `HLC` is supported
- `--symbol` now just `HLC` is supported
- `--i` Intensities (the work size is 2^intensity) up to device
- `--W` The explicitly declared sizes of the work to do up to device (overrides intensity)
- `--mineraddress` the miner account address
- `-o` the pool address
- `-m` the pool user account address
- `--notls` rpc not use tls
- `--rpccert` rpc use tls with cert path,can get from node
- `--pow` the type of pow `blake2bd|cuckaroo|cuckatoo`
- `--trimmerCount` trimmer times ,default 40

## Supported coin 
        
  - `HLC`

## Supported pow 
        
  - `blake2bd` double blake2b
  - `cuckaroo` 
  - `cuckatoo`  todo
        
## Directory structure

- common  the universally function
    
- core every coin miner must realize :
    1) devices to mining the result
    2) robot handle the logic
    3) work get or submit the task 
    
- symbols 
    
    - like `HLC`
