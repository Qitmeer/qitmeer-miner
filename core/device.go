/**
	HLC FOUNDATION
	james
 */
package core

import (
	"github.com/robvanmieghem/go-opencl/cl"
	"os"
	"time"
	"sync"
	"hlc-miner/common"
	"log"
)
type Device struct{
	Cfg *common.Config  //must init
	DeviceName string
	HasNewWork bool
	AllDiffOneShares uint64
	AverageHashRate    float64
	MinerId          uint32
	Context          *cl.Context
	CommandQueue     *cl.CommandQueue
	LocalItemSize     int
	NonceOut     []byte
	BlockObj     *cl.MemObject
	NonceOutObj     *cl.MemObject
	Kernel     *cl.Kernel
	Program     	*cl.Program
	ClDevice         *cl.Device
	Started          uint32
	GlobalItemSize int
	CurrentWorkID uint32
	Quit chan os.Signal //must init
	sync.Mutex
	Wg sync.WaitGroup
	Pool bool //must init
	IsValid bool //is valid
	SubmitData chan string //must
}

func (this *Device)Mine()  {
}

func (this *Device)Update()  {
	this.CurrentWorkID++
}

func (this *Device)InitDevice()  {
	var err error
	this.Context, err = cl.CreateContext([]*cl.Device{this.ClDevice})
	if err != nil {
		this.IsValid = false
		log.Println("-", this.MinerId, err)
		return
	}
	this.CommandQueue, err = this.Context.CreateCommandQueue(this.ClDevice, 0)
	if err != nil {
		this.IsValid = false
		log.Println("-", this.MinerId,  err)
	}
}

func (d *Device)Release()  {
	d.Kernel.Release()
	d.Context.Release()
	d.BlockObj.Release()
	d.NonceOutObj.Release()
	d.Program.Release()
	d.CommandQueue.Release()
}

func (this *Device)Status()  {
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	for {

		select{
		case <- this.Quit:
			return
		case <- t.C:
			if !this.IsValid{
				return
			}
			secondsElapsed := uint32(time.Now().Unix()) - this.Started
			diffOneShareHashesAvg := uint64(0x00000000FFFFFFFF)
			averageHashRate := (float64(diffOneShareHashesAvg) *
				float64(this.AllDiffOneShares)) /
				float64(secondsElapsed)
			log.Printf("DEVICE_ID #%d (%s) %v",
				this.MinerId,
				this.ClDevice.Name(),
				common.FormatHashRate(averageHashRate),
			)
		}
	}
}