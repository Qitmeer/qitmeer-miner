package main

import (
	`fmt`
	`github.com/Qitmeer/go-opencl/cl`
)

func main()  {
	platforms, err := cl.GetPlatforms()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, platform := range platforms {
		
		platormDevices, err := cl.GetDevices(platform, cl.DeviceTypeGPU)
		if err != nil {
			continue
		}
		for _, device := range platormDevices {
			fmt.Println(device.Name())
		}
	}
}
