/**
Qitmeer
james
*/
package core

import (
	`fmt`
	"github.com/Qitmeer/go-opencl/cl"
	"math"
	"os"
	"qitmeer-miner/common"
	"sync"
	"time"
)

type BaseDevice interface {
	Mine(wg *sync.WaitGroup)
	Update()
	InitDevice()
	Status(wg *sync.WaitGroup)
	GetIsValid() bool
	SetIsValid(valid bool)
	GetMinerId() int
	GetAverageHashRate() float64
	GetName() string
	GetStart() uint64
	GetIntensity() int
	GetWorkSize() int
	SetIntensity(inter int)
	SetWorkSize(lsize int)
	SetPool(pool bool)
	SetNewWork(w BaseWork)
	SetForceUpdate()
	Release()
	GetMinerType() string
	SubmitShare(substr chan string)
	StopTask()
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
	NonceRandObj     *cl.MemObject
	Target2Obj     *cl.MemObject
	Kernel     *cl.Kernel
	Program     	*cl.Program
	ClDevice         *cl.Device
	Started          int64
	GlobalItemSize int
	CurrentWorkID uint64
	Quit chan os.Signal //must init
	sync.Mutex
	Wg         sync.WaitGroup
	Pool       bool //must init
	IsValid    bool //is valid
	SubmitData chan string //must
	NewWork    chan BaseWork
	Err        error
	MiningType        string
	Event *cl.Event
	StopTaskChan chan bool
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
	this.SubmitData=make(chan string,1)
	this.GlobalItemSize= int(math.Exp2(float64(this.Cfg.OptionConfig.Intensity)))
	this.Quit=q
	this.AllDiffOneShares = 0
	this.StopTaskChan = make(chan bool,1)
}

func (this *Device)GetIsValid() bool {
	return this.IsValid
}

func (this *Device)SetNewWork(work BaseWork) {
	if !this.GetIsValid(){
		return
	}
	this.HasNewWork = true
	this.NewWork <- work
}

func (this *Device)StopTask() {
	if !this.GetIsValid(){
		return
	}
	this.StopTaskChan <- true
}

func (this *Device)SetForceUpdate() {
	if !this.GetIsValid(){
		return
	}
	this.HasNewWork = true
	this.AllDiffOneShares = 0
}


func (this *Device)GetMinerType() string{
	return this.MiningType
}

func (this *Device)Update()  {
	this.CurrentWorkID = common.RandUint64()
}

func (this *Device)InitDevice()  {
	var err error
	this.Context, err = cl.CreateContext([]*cl.Device{this.ClDevice})
	if err != nil {
		this.IsValid = false
		common.MinerLoger.Info("CreateContext", "minerId",this.MinerId,"error", err)
		return
	}
	this.CommandQueue, err = this.Context.CreateCommandQueue(this.ClDevice, 0)
	if err != nil {
		this.IsValid = false
		common.MinerLoger.Info("CreateCommandQueue","minerId", this.MinerId, "error", err)
	}
}

func (this *Device)SetPool(b bool)  {
	this.Pool = b
}

func (this *Device)SetIsValid(valid bool) {
	this.IsValid = valid
}

func (this *Device)GetMinerId() int {
	return int(this.MinerId)
}

func (this *Device)GetIntensity() int {
	return int(math.Log2(float64(this.GlobalItemSize)))
}

func (this *Device)GetWorkSize() int {
	return this.LocalItemSize
}

func (this *Device)SetIntensity(inter int) {
	this.GlobalItemSize = int(math.Exp2(float64(this.Cfg.OptionConfig.Intensity)))
}

func (this *Device)SetWorkSize(size int) {
	this.LocalItemSize = size
}

func (this *Device)GetName() string {
	return this.DeviceName
}

func (this *Device)GetStart() uint64 {
	return uint64(this.Started)
}

func (this *Device)GetAverageHashRate() float64 {
	return this.AverageHashRate
}

func (d *Device)Release()  {
	d.Kernel.Release()
	d.Context.Release()
	d.BlockObj.Release()
	d.NonceOutObj.Release()
	d.Program.Release()
	d.NonceRandObj.Release()
	d.Target2Obj.Release()
	d.CommandQueue.Release()
}

func (this *Device)Status(wg *sync.WaitGroup)  {
	defer wg.Done()
	for {
		select{
		case <- this.Quit:
			return
		default:
			common.Usleep(10*1000)
			if !this.IsValid{
				return
			}
			secondsElapsed := time.Now().Unix() - this.Started
			//diffOneShareHashesAvg := uint64(0x00000000FFFFFFFF)
			if this.AllDiffOneShares <= 0 || secondsElapsed <= 0{
				common.Usleep(10*1000)
				continue
			}
			averageHashRate := float64(this.AllDiffOneShares) /
				float64(secondsElapsed)
			if averageHashRate <= 0{
				common.Usleep(10*1000)
				continue
			}
			if this.AverageHashRate <= 0{
				this.AverageHashRate = averageHashRate
			}
			//recent stats 95% percent
			this.AverageHashRate = (this.AverageHashRate*50+averageHashRate*950)/1000
			unit := " H/s"
			if this.GetMinerType() != "blake2bd"{
				unit = " GPS"
			}
			common.MinerLoger.Info(fmt.Sprintf("# %d [%s] : %s",this.MinerId,this.ClDevice.Name(),common.FormatHashRate(this.AverageHashRate,unit)))
		}
	}
}

func (this *Device) SubmitShare(substr chan string) {
	if !this.GetIsValid(){
		return
	}
	for {
		common.MinerLoger.Debug("===============================Listen Submit")
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
