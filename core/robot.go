// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package core

import (
	"github.com/Qitmeer/go-opencl/cl"
	"qitmeer-miner/common"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	SYMBOL_PMEER = "PMEER"
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
	clDs := common.GetDevices(typ)
	if clDs == nil{
		common.MinerLoger.Infof("some error occurs!")
		return
	}
	useDevices := []string{}
	if this.Cfg.OptionConfig.UseDevices != ""{
		useDevices = strings.Split(this.Cfg.OptionConfig.UseDevices,",")
	}
	if len(useDevices) > 0{
		for k := range clDs{
			if common.InArray(strconv.Itoa(k),useDevices){
				common.MinerLoger.Infof("【Select mining Devices】",k,clDs[k].Name())
				this.ClDevices = append(this.ClDevices,clDs[k])
			}
		}
	} else{
		this.ClDevices = clDs
	}
	if len(this.ClDevices) < 1{
		log.Fatalln("You Don't select any GPU devices to mining!")
		return
	}
}
