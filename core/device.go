/**
Qitmeer
james
*/
package core

import (
	"github.com/Qitmeer/go-opencl/cl"
	"log"
	"math"
	"os"
	"qitmeer-miner/common"
	"sync"
	"time"
)

type BaseDevice interface {
	Mine()
	Update()
	InitDevice()
	Status()
	GetIsValid() bool
	SetNewWork(w BaseWork)
	Release()
	SubmitShare(substr chan string)
}
type Device struct{
	Cfg *common.GlobalConfig  //must init
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
	Started          int64
	GlobalItemSize int
	CurrentWorkID uint64
	Quit chan os.Signal //must init
	sync.Mutex
	Wg sync.WaitGroup
	Pool bool //must init
	IsValid bool //is valid
	SubmitData chan string //must
	NewWork chan BaseWork
}

func (this *Device)Init(i int,device *cl.Device,pool bool,q chan os.Signal,cfg *common.GlobalConfig)  {
	this.MinerId = uint32(i)
	this.NewWork = make(chan BaseWork,1)
	this.Cfg=cfg
	this.DeviceName=device.Name()
	this.ClDevice=device
	this.CurrentWorkID=0
	this.IsValid=true
	this.Pool=pool
	this.SubmitData=make(chan string)
	this.Started=time.Now().Unix()
	this.GlobalItemSize= int(math.Exp2(float64(this.Cfg.OptionConfig.Intensity)))
	this.Quit=q
	this.AllDiffOneShares = 0
}

func (this *Device)Mine()  {
}

func (this *Device)Update()  {
	defer func() {
		err := recover()
		if err != nil {
			log.Println("[error]",err)
		}
	}()
	var err error
	this.CurrentWorkID,err = common.RandUint64()
	if err != nil{
		this.CurrentWorkID++
	}
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

func (this *Device)SetNewWork(w BaseWork)  {
	this.HasNewWork = true
	this.NewWork <- w
}

func (this *Device)GetIsValid() bool {
	return this.IsValid
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
			secondsElapsed := time.Now().Unix() - this.Started
			//diffOneShareHashesAvg := uint64(0x00000000FFFFFFFF)
			if this.AllDiffOneShares <= 0 || secondsElapsed <= 0{
				continue
			}
			averageHashRate := float64(this.AllDiffOneShares) /
				float64(secondsElapsed)
			this.AverageHashRate = (this.AverageHashRate+averageHashRate)/2
			log.Printf("DEVICE_ID #%d (%s) %v",
				this.MinerId,
				this.ClDevice.Name(),
				common.FormatHashRate(this.AverageHashRate),
			)
		}
	}
}

func (this *Device) SubmitShare(substr chan string) {
	for {
		select {
		case <-this.Quit:
			return
		case str := <-this.SubmitData:
			if this.HasNewWork {
				//the stale submit
				continue
			}
			substr <- str
		}
	}
}
