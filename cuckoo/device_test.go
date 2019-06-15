package cuckoo

import (
	"fmt"
	"github.com/robvanmieghem/go-opencl/cl"
	"log"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T){
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


	platforms, err := cl.GetPlatforms()
	if err != nil {
		log.Fatalln("Get Graphics card platforms error,please check!【",err,"】")
		return
	}
	clDevices := make([]*cl.Device, 0)

	for _, platform := range platforms {
		platormDevices, err := cl.GetDevices(platform, cl.DeviceTypeGPU)
		if err != nil {
			log.Fatalln("Don't had GPU devices to mining ,please check!【",err,"】")
			return
		}
		for _, device := range platormDevices {
			if device.Type() != cl.DeviceTypeGPU {
				continue
			}
			log.Println(device.Name())
			if strings.Contains(device.Name(),"Graphics"){
				continue
			}

			clDevices = append(clDevices, device)
		}
	}

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

	for _,this := range devices{
		this.Mine()
	}

}