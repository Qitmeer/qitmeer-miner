package cuckoo

import (
	"github.com/robvanmieghem/go-opencl/cl"
	"hlc-miner/kernel"
	"os"
	"sync"
	"log"
)
type Device struct{
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
}

func (this *Device)Mine()  {
	for{
		offset := 0xffffffff
		for {
			var err error
			headerD := make([]byte,100)
			if _, err = this.CommandQueue.EnqueueWriteBufferByte(this.BlockObj, true, 0, headerD, nil); err != nil {
				log.Println("-", this.MinerId,err)
				return
			}
			//Run the kernel
			if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernel, []int{int(offset)}, []int{this.GlobalItemSize}, []int{this.LocalItemSize}, nil); err != nil {
				log.Println("-", this.MinerId,err)
				return
			}
			offset--
			//Get output
			if _, err = this.CommandQueue.EnqueueReadBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); err != nil {
				log.Println("-", this.MinerId,err)
				return
			}
			if this.NonceOut[0] != 0 || this.NonceOut[1] != 0 || this.NonceOut[2] != 0 || this.NonceOut[3] != 0 ||
				this.NonceOut[4] != 0 || this.NonceOut[5] != 0 || this.NonceOut[6] != 0 || this.NonceOut[7] != 0 {
				//Found Hash
			}
		}
	}
}

func (this *Device)Update()  {
	this.CurrentWorkID++
}

func (this *Device)InitDevice()  {
	var err error
	this.Context, err = cl.CreateContext([]*cl.Device{this.ClDevice})
	if err != nil {
		log.Println("-", this.MinerId, err)
		return
	}
	this.CommandQueue, err = this.Context.CreateCommandQueue(this.ClDevice, 0)
	if err != nil {
		log.Println("-", this.MinerId,  err)
	}
	this.Program, err = this.Context.CreateProgramWithSource([]string{kernel.CuckarooKernel})
	if err != nil {
		log.Println("-", this.MinerId, err)
		return
	}

	err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
	if err != nil {
		log.Println("-", this.MinerId, err)
		return
	}

	this.Kernel, err = this.Program.CreateKernel("search")
	if err != nil {
		log.Println("-", this.MinerId, err)
		return
	}
	this.BlockObj, err = this.Context.CreateEmptyBuffer(cl.MemReadOnly, 128)
	if err != nil {
		log.Println("-", this.MinerId, err)
		return
	}
	this.Kernel.SetArgBuffer(0, this.BlockObj)
	this.NonceOutObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, 8)
	if err != nil {
		log.Println("-", this.MinerId, err)
		return
	}
	this.Kernel.SetArgBuffer(1, this.NonceOutObj)
	this.LocalItemSize, err = this.Kernel.WorkGroupSize(this.ClDevice)
	this.LocalItemSize = 256
	if err != nil {
		log.Println("- WorkGroupSize failed -",this.MinerId, err)
		return
	}
	log.Println("- Device ID:", this.MinerId, "- Global item size:",this.GlobalItemSize, "(Intensity", 24, ")", "- Local item size:", this.LocalItemSize)
	this.NonceOut = make([]byte, 8, 8)
	if _, err = this.CommandQueue.EnqueueWriteBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); err != nil {
		log.Println("-",this.MinerId, err)
		return
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
