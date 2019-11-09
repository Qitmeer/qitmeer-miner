---
name: qitmeer-miner软件BUG报告 
about: 提交BUG以帮助我们改善挖矿软件的质量和使用体验 
---

**BUG描述**
请用简明扼要的语音描述该BUG

**重现步骤**
如何可以使该BUG出现的步骤:
1. 首先 "..."
2. 第2步, "..."
3. 第3步,  "..."
4. 可以看到错误如下...

**预期的行为**
请用简明扼要的语音描述正确的行为是什么

**截屏(如有)**
如果有，可以使用截屏来帮助描述问题所在。

**环境信息 (请尽可能提供完整信息):**
- 操作系统 [e.g. Windows 10]
- 显卡 [e.g. GTX 1070]
  你可以使用`./qitmeer-miner -l`命令获得输出粘贴在这里，如:
  ```
  $ ./qitmeer-miner -l
  2019-11-09|14:06:04.344 [INFO ] [CPU Devices List]:                 module=miner
  2019-11-09|14:06:04.347 [INFO ] Found Device                        module=miner platform=Apple minerID=0 deviceName="Intel(R) Core(TM) i7-7920HQ CPU @ 3.10GHz" MaxWorkGroupSize(MB)=1024 MaxMemAllocSize(MB)=4096.000
  2019-11-09|14:06:04.347 [INFO ] [GPU Devices List]:                 module=miner
  2019-11-09|14:06:04.380 [INFO ] Found Device                        module=miner platform=Apple minerID=0 deviceName="Intel(R) HD Graphics 630"                  MaxWorkGroupSize(MB)=256  MaxMemAllocSize(MB)=384.000
  2019-11-09|14:06:04.380 [INFO ] Found Device                        module=miner platform=Apple minerID=1 deviceName="AMD Radeon Pro 560 Compute Engine"         MaxWorkGroupSize(MB)=256  MaxMemAllocSize(MB)=1024.000
  ```

- qitmeer-miner 版本 [e.g. 0.2.4]
  你可以使用`./qitmeer-miner --version`命令获得输出粘贴在这里，如:
  ```
  $ ./qitmeer-miner --version
  Qitmeer Miner Version:0.2.4
  ```

- 命令行参数或配置文件 
  例如：
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

**其它备注**
任何其它需要说明的问题。
