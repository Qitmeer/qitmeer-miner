package common

import (
	"github.com/Qitmeer/go-opencl/cl"
)

var DevicesTypesForGPUMining = cl.DeviceTypeGPU
var DevicesTypesForCPUMining = cl.DeviceTypeCPU
func GetDevices(t cl.DeviceType) []*cl.Device {
	platforms, err := cl.GetPlatforms()
	if err != nil {
		MinerLoger.Errorf("Get Graphics card platforms error,please check![%s]",err.Error())
		return nil
	}
	clDevices := make([]*cl.Device, 0)
	i := 0
	for _, platform := range platforms {
		platormDevices, err := cl.GetDevices(platform, t)
		if err != nil {
			MinerLoger.Errorf("%s Get Devices Error:%s",platform.Name(),err.Error())
			continue
		}
		for _, device := range platormDevices {
			clDevices = append(clDevices, device)
			MinerLoger.Infof("%s Found Device : %d | name: %s | MaxGroupSize: %d | MaxAllocMemory: %.2f MB | MaxMemory: %0.2f",platform.Name(),i,device.Name(),
				device.MaxWorkGroupSize(),float64(device.MaxMemAllocSize())/1024.00/1024.00,float64(device.GlobalMemSize())/1024.00/1024.00 )

			i++
		}
	}
	if len(clDevices) < 1{
		MinerLoger.Error("Don't had devices to mining,please check your PC!")
		return nil
	}
	return clDevices
}