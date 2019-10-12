package cuckoo

import (
    `fmt`
    go_logger `github.com/phachon/go-logger`
    `math`
    `os`
    `qitmeer-miner/common`
    "testing"
    `time`
)
//init the config file
func init(){
    common.MinerLoger = go_logger.NewLogger()
}

func TestCuckoo29(t *testing.T){
    err := os.Setenv("GPU_MAX_HEAP_SIZE", "100")
    if err != nil {
        fmt.Println(err.Error())
    }
    err = os.Setenv("GPU_USE_SYNC_OBJECTS", "1")
    if err != nil {
        fmt.Println(err.Error())
    }
    err = os.Setenv("GPU_MAX_ALLOC_PERCENT", "95")
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


    clDevices := common.GetDevices(common.DevicesTypesForGPUMining)

    devices := make([]*Device,0)

    for i, device := range clDevices {
        deviceMiner := &Device{
        }
        deviceMiner.MinerId = uint32(i)
        deviceMiner.DeviceName=device.Name()
        deviceMiner.ClDevice=device
        deviceMiner.CurrentWorkID=0
        deviceMiner.Started=uint32(time.Now().Unix())
        deviceMiner.GlobalItemSize= int(math.Exp2(float64(24)))
        devices = append(devices,deviceMiner)
    }

    for k,d := range devices{
        if k == 0{
            continue
        }
        d.Mine()
        break
    }
}