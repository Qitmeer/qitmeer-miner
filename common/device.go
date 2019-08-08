package common

import (
	"fmt"
	"github.com/HalalChain/go-opencl/cl"
	"log"
)

var DevicesTypesForGPUMining = cl.DeviceTypeGPU
var DevicesTypesForCPUMining = cl.DeviceTypeCPU
func GetDevices(t cl.DeviceType) []*cl.Device {
	platforms, err := cl.GetPlatforms()
	if err != nil {
		log.Fatalln("Get Graphics card platforms error,please check!【",err,"】")
		return nil
	}
	clDevices := make([]*cl.Device, 0)
	i := 0
	for _, platform := range platforms {
		log.Println(platform.Name())
		platormDevices, err := cl.GetDevices(platform, t)
		if err != nil {
			log.Println(platform.Name(),"Don't had Any GPU devices!")
			continue
		}
		for _, device := range platormDevices {
			clDevices = append(clDevices, device)
			log.Println(platform.Name(),fmt.Sprintf("Found Device : %d | name: %s | MaxGroupSize: %d | MaxAllocMemory: %.2f MB | MaxGlobalMemory: %.2f MB",i,device.Name(),
				device.MaxWorkGroupSize(),float64(device.MaxMemAllocSize())/1024.00/1024.00 ,float64(device.GlobalMemSize())/1024.00/1024.00 ) )

			i++
		}
	}
	if len(clDevices) < 1{
		log.Fatalln("Don't had GPU devices to mining,please check your PC!")
		return nil
	}
	return clDevices
}