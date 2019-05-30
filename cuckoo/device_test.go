package cuckoo

import (
	"github.com/robvanmieghem/go-opencl/cl"
	"log"
	"math"
	"testing"
	"time"
)

func TestRun(t *testing.T){

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
		this.Context, err = cl.CreateContext([]*cl.Device{this.ClDevice})
		if err != nil {
			log.Println("-1", this.MinerId, err)
			return
		}
		this.CommandQueue, err = this.Context.CreateCommandQueue(this.ClDevice, 0)
		if err != nil {
			log.Println("-2", this.MinerId,  err)
		}
		this.Program, err = this.Context.CreateProgramWithSource([]string{CuckarooKernel})
		if err != nil {
			log.Println("-3", this.MinerId, this.DeviceName, err)
			return
		}

		err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
		if err != nil {
			log.Println("-build", this.MinerId, err)
			return
		}
		this.Mine()
	}

}