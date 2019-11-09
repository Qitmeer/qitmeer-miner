---
name: qitmeer-miner Bug report
about: Please submit bug reprot to help us improve software quality and user experience 
---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to "..."
2. Click on "..."
3. Scroll down to "..."
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots (Optional)**
If applicable, add screenshots to help explain your problem.

**Environment (please complete the following information):**
- Operating System: [e.g. Windows 10]
- Hardware [e.g. GTX 1070]
  you can paste the output of `./qitmeer-miner -l` e.g:
  ```
  $ ./qitmeer-miner -l
  2019-11-09|14:06:04.344 [INFO ] [CPU Devices List]:                 module=miner
  2019-11-09|14:06:04.347 [INFO ] Found Device                        module=miner platform=Apple minerID=0 deviceName="Intel(R) Core(TM) i7-7920HQ CPU @ 3.10GHz" MaxWorkGroupSize(MB)=1024 MaxMemAllocSize(MB)=4096.000
  2019-11-09|14:06:04.347 [INFO ] [GPU Devices List]:                 module=miner
  2019-11-09|14:06:04.380 [INFO ] Found Device                        module=miner platform=Apple minerID=0 deviceName="Intel(R) HD Graphics 630"                  MaxWorkGroupSize(MB)=256  MaxMemAllocSize(MB)=384.000
  2019-11-09|14:06:04.380 [INFO ] Found Device                        module=miner platform=Apple minerID=1 deviceName="AMD Radeon Pro 560 Compute Engine"         MaxWorkGroupSize(MB)=256  MaxMemAllocSize(MB)=1024.000
  ```
- qitmeer-miner Version [e.g. 0.2.4]
  you can paste the output of `./qitmeer-miner --version` e.g:
  ```
  $ ./qitmeer-miner --version
  Qitmeer Miner Version:0.2.4
  ```
- qitmeer-miner options or `.conf` file 
  eg :
  ```
  #### Device Config ####
  # all gpu devices,you can use ./qitmeer-miner -l to see. examples:0,1 use the #0 device and #1 device
  use_devices=1

  #### Log Config ####
  # specify a file to write miner log
  #minerlog=
  # log level : info|debug|error|warn|trace
  log_level=debug

  #### Cuckoo Config ####
  # edge bits (24)
  edge_bits=24
  # the cuckaroo trimmer times (15)
  # trimmerTimes can ajustment this parameter to keep performance
  trimmerTimes=15

  #### Blake2bd Config ####
  # Intensities (the work size is 2^intensity) up to device
  #intensity=24
  # The explicitly declared sizes of the work to do per device (overrides intensity). Single global value or a comma separated list. (256)
  # worksize=256
  # work group size (256)
  #group_size=

  #### Other Config ####
  #GPU local size (4096)
  #local_size=
  ##rpc timeout. (60)
  #timeout=60
  #max pack tx count (1000)
  #max_tx_count=
  #max sign tx count (4000)
  max_sig_count=4000
  #stats web server (127.0.0.1:1235)
  #stats_server=
  ```

**Additional context**
Add any other context about the problem here.
