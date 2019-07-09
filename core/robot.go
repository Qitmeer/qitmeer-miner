// Copyright (c) 2019 The halalchain developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package core

import (
	"github.com/HalalChain/go-opencl/cl"
	"hlc-miner/common"
	"log"
	"os"
	"sync"
)

const (
	SYMBOL_NOX = "HLC"
)

//var devicesTypesForMining = cl.DeviceTypeAll

type Robot interface {
	Run()	// uses device to calulate the nonce
	ListenWork() //listen the solo or pool work
	SubmitWork() //submit the work
}

type MinerRobot struct {
	Cfg 		  *common.GlobalConfig //config
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
	var typ = common.DevicesTypesForGPUMining
	if this.Cfg.OptionConfig.CPUMiner{
		typ = common.DevicesTypesForCPUMining
	}
	this.ClDevices = common.GetDevices(typ)
	if this.ClDevices == nil{
		log.Println("some error occurs!")
		return
	}
}
