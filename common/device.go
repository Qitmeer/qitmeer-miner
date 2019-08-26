package common

import (
	"fmt"
	"github.com/Qitmeer/go-opencl/cl"
	"log"
)

var DevicesTypesForGPUMining = cl.DeviceTypeGPU
var DevicesTypesForCPUMining = cl.DeviceTypeCPU
func GetDevices(t cl.DeviceType) []*cl.Device {
	platforms, err := cl.GetPlatforms()
	if err != nil {
		log.Println("Get Graphics card platforms error,please check!【",err.Error(),"】")
		return nil
	}
	clDevices := make([]*cl.Device, 0)
	i := 0
	for _, platform := range platforms {
		platormDevices, err := cl.GetDevices(platform, t)
		if err != nil {
			log.Println(platform.Name(),"Get Devices Error:",err.Error())
			continue
		}
		for _, device := range platormDevices {
			clDevices = append(clDevices, device)
			log.Println(platform.Name(),fmt.Sprintf("Found Device : %d | name: %s | MaxGroupSize: %d | MaxAllocMemory: %.2f MB",i,device.Name(),
				device.MaxWorkGroupSize(),float64(device.MaxMemAllocSize())/1024.00/1024.00 ))

			i++
		}
	}
	if len(clDevices) < 1{
		log.Println("Don't had devices to mining,please check your PC!")
		return nil
	}
	return clDevices
}