/**
	HLC FOUNDATION
	james
 */
package core

import (
	"sync"
	"os"
	"github.com/robvanmieghem/go-opencl/cl"
	"log"
	"hlc-miner/common"
)

const (
	SYMBOL_NOX = "HLC"
)
var devicesTypesForMining = cl.DeviceTypeGPU
//var devicesTypesForMining = cl.DeviceTypeAll

type Robot interface {
	Run()	// uses device to calulate the nonce
	ListenWork() //listen the solo or pool work
	SubmitWork() //submit the work
}

type MinerRobot struct {
	Cfg 		  *common.Config //config
	ValidShares   uint64
	StaleShares   uint64
	InvalidShares uint64
	AllDiffOneShares uint64
	Wg            sync.WaitGroup
	Started       uint32
	Quit          chan os.Signal
	Work 		  *Work
	ClDevices 	  []*cl.Device
	Rpc 		*common.RpcClient
	Pool 		bool
	SubmitStr chan string
}

//init GPU device
func (this *MinerRobot)InitDevice()  {
	platforms, err := cl.GetPlatforms()
	if err != nil {
		log.Fatalln("Get Graphics card platforms error,please check!【",err,"】")
		return
	}
	this.ClDevices = make([]*cl.Device, 0)

	for _, platform := range platforms {
		platormDevices, err := cl.GetDevices(platform, devicesTypesForMining)
		if err != nil {
			log.Fatalln("Don't had GPU devices to mining ,please check!【",err,"】")
			return
		}
		for _, device := range platormDevices {
			if device.Type() != cl.DeviceTypeGPU {
				continue
			}
			this.ClDevices = append(this.ClDevices, device)
			//break
		}
		//break
	}

}
