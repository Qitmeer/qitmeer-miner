package main

import (
    `fmt`
    go_logger `github.com/phachon/go-logger`
    `math`
    `os`
    `qitmeer-miner/common`
    `qitmeer-miner/cuckoo`
    `sync`
    `time`
)
//init the config file
func init(){
    common.MinerLoger = go_logger.NewLogger()
}

func main() {
    err := os.Setenv("GPU_MAX_HEAP_SIZE", "100")
    if err != nil {
        fmt.Println(err.Error())
    }
    err = os.Setenv("GPU_USE_SYNC_OBJECTS", "1")
    if err != nil {
        fmt.Println(err.Error())
    }
    err = os.Setenv("GPU_MAX_ALLOC_PERCENT", "100")
    if err != nil {
        fmt.Println(err.Error())
    }
    err = os.Setenv("GPU_SINGLE_ALLOC_PERCENT", "100")
    if err != nil {
        fmt.Println(err.Error())
    }
    err = os.Setenv("GPU_64BIT_ATOMICS", "100")
    if err != nil {
        fmt.Println(err.Error())
    }
    err = os.Setenv("GPU_FORCE_64BIT_PTR", "100")
    if err != nil {
        fmt.Println(err.Error())
    }
    err = os.Setenv("GPU_MAX_WORKGROUP_SIZE", "1024")
    if err != nil {
        fmt.Println(err.Error())
    }
    err = os.Setenv("CL_LOG_ERRORS", "stdout")
    if err != nil {
        common.MinerLoger.Errorf(err.Error())
        return
    }
    clDevices := common.GetDevices(common.DevicesTypesForGPUMining)

    devices := make([]*cuckoo.Device,0)

    for i, device := range clDevices {
        deviceMiner := &cuckoo.Device{
        }
        deviceMiner.MinerId = uint32(i)
        deviceMiner.DeviceName=device.Name()
        deviceMiner.ClDevice=device
        deviceMiner.CurrentWorkID=0
        deviceMiner.Started=time.Now().Unix()
        deviceMiner.GlobalItemSize= int(math.Exp2(float64(24)))
        devices = append(devices,deviceMiner)
    }
    wg := sync.WaitGroup{}
    for k,d := range devices{
        if k == 0 {
            continue
        }
        common.MinerLoger.Debugf("允许单对象最大 %d G",d.ClDevice.MaxMemAllocSize()/1000/1000/1000)
        common.MinerLoger.Debugf("允许内存最大 %d G",d.ClDevice.GlobalMemSize()/1000/1000/1000)
        wg.Add(1)
        go d.Status(&wg)
        d.SetIsValid(true)
        d.InitDevice()
        d.Mine()
        break
    }
    wg.Wait()
}