# HLC Miner

    The miner of Halalchain

## Enviroment

    go version >= 1.11
    go mod
    replace
        qitmeer v0.0.0-20190510071513-7cff93db4878 => github.com/HalalChain/qitmeer
        or git clone git@github.com:HalalChain/qitmeer.git in current directory
## Compile

* Ubuntu ENV

        sudo apt-get install beignet-dev nvidia-cuda-dev nvidia-cuda-toolkit
        go build 
* Centos ENV

        sudo yum install opencl-headers
        sudo yum install ocl-icd
        sudo ln -s /usr/lib64/libOpenCL.so.1 /usr/lib/libOpenCL.so
        go build

* MAC

        go build
    
* Windows ENV

        install the opencl driver
    
        go build 
    
## Run

    cp halalchainminer.conf.example halalchainminer.conf
    ./hlc-miner
    
    command params mode:
        solo

            Network:
                simnet
                testnet
                mainnet

            ./hlc-miner -s 127.0.0.1:1234 -u test -P test --symbol HLC --notls -i 24 -W 256 --mineraddress RmN4SADy42FKmN8ARKieX9iHh9icptdgYNn --simnet


         pool
            ./hlc-miner -o stratum+tcp://127.0.0.1:3177 -m RmN4SADy42FKmN8ARKieX9iHh9icptdgYNn --symbol HLC --notls -i 24 -W 256
            
            -i Intensities (the work size is 2^intensity) per device
            -W The explicitly declared sizes of the work to do per device (overrides intensity)
    
        dag miner 
            add tag
            --dag
## Supported coin 
        
        HLC
        
## Directory structure

    common  the universally function
    
    core every coin miner must realize :
    1) devices to mining the result
    2) robot handle the logic
    3) work get or submit the task 
    
    symbols 
        look HLC 
