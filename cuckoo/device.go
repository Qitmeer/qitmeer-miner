package cuckoo

import (
	"github.com/robvanmieghem/go-opencl/cl"
	"os"
	"sync"
	"log"
	"encoding/binary"
)
const DUCK_SIZE_A  = 129 // AMD 126 + 3
const DUCK_SIZE_B  = 83
const BUFFER_SIZE_A1  = DUCK_SIZE_A * 1024 * (4096 - 128) * 2
const BUFFER_SIZE_A2 = DUCK_SIZE_A * 1024 * 256 * 2
const BUFFER_SIZE_B = DUCK_SIZE_B * 1024 * 4096 * 2
const INDEX_SIZE = 256 * 256 * 4
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
	BufferA     *cl.MemObject
	BufferB     *cl.MemObject
	Indexes     *cl.MemObject
	NonceOutObj     *cl.MemObject
	Kernels     map[string]*cl.Kernel
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
	var err error
	var k = []uint64{
		0xf4956dc403730b01,
		0xe6d45de39c2a5a3e,
		0xcbf626a8afee35f6,
		0x4307b94b1a0c9980,
	}

	//init kernels
	this.Kernels = make(map[string]*cl.Kernel,0)
	this.Kernels["seedA"], _ = this.Program.CreateKernel("FluffySeed2A")
	this.Kernels["seedA"].SetArg(0,k[0])
	this.Kernels["seedA"].SetArg(1,k[1])
	this.Kernels["seedA"].SetArg(2,k[2])
	this.Kernels["seedA"].SetArg(3,k[3])
	A1, _ := this.Context.CreateEmptyBuffer(cl.MemReadOnly, BUFFER_SIZE_A1)
	A2, _ := this.Context.CreateEmptyBuffer(cl.MemReadOnly, BUFFER_SIZE_A2)
	B, _ := this.Context.CreateEmptyBuffer(cl.MemReadOnly, BUFFER_SIZE_B)
	I1, _ := this.Context.CreateEmptyBuffer(cl.MemReadOnly, INDEX_SIZE)
	I2, _ := this.Context.CreateEmptyBuffer(cl.MemReadOnly, INDEX_SIZE)
	//R ,_:= this.Context.CreateEmptyBuffer(cl.MemReadOnly, 42*2);
	this.Kernels["seedA"].SetArgBuffer(4,B)
	this.Kernels["seedA"].SetArgBuffer(5,A1)
	this.Kernels["seedA"].SetArgBuffer(6,I1)

	b := make([]byte,BUFFER_SIZE_B)
	_,err = this.CommandQueue.EnqueueWriteBufferByte(B,true,0,b,nil)
	log.Println(err)
	//a1 := make([]byte,BUFFER_SIZE_A1)
	//_,err = this.CommandQueue.EnqueueWriteBufferByte(A1,true,0,a1,nil)
	//log.Println(err)
	//c1 := make([]byte,INDEX_SIZE)
	//_,err = this.CommandQueue.EnqueueWriteBufferByte(I1,true,0,c1,nil)
	//log.Println(err)
	if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["seedA"], []int{0}, []int{2048*128}, []int{128}, nil); err != nil {
		log.Println("-123", this.MinerId,err)
		return
	}
	log.Println(12344)
	return
	edges_left := make([]byte,INDEX_SIZE)
	//Get output
	if _, err = this.CommandQueue.EnqueueReadBufferByte(I1, true, 0, edges_left, nil); err != nil {
		log.Println("- ReadBuffer", this.MinerId,err)
		return
	}
	log.Println("======",edges_left)
	return
	this.Kernels["seedB1"], err = this.Program.CreateKernel("FluffySeed2B")
	this.Kernels["seedB1"].SetArg(0,A1)
	this.Kernels["seedB1"].SetArg(1,A1)
	this.Kernels["seedB1"].SetArg(2,A2)
	this.Kernels["seedB1"].SetArg(3,I1)
	this.Kernels["seedB1"].SetArg(4,I2)
	this.Kernels["seedB1"].SetArg(5,uint32(32))
	this.Kernels["seedB2"], err = this.Program.CreateKernel("FluffySeed2B")
	this.Kernels["seedB2"].SetArg(0,B)
	this.Kernels["seedB2"].SetArg(1,A1)
	this.Kernels["seedB2"].SetArg(2,A2)
	this.Kernels["seedB2"].SetArg(3,I1)
	this.Kernels["seedB2"].SetArg(4,I2)
	this.Kernels["seedB2"].SetArg(5,uint32(0))
	this.Kernels["round1"], err = this.Program.CreateKernel("FluffyRound1")
	this.Kernels["round1"].SetArg(0,A1)
	this.Kernels["round1"].SetArg(1,A2)
	this.Kernels["round1"].SetArg(2,B)
	this.Kernels["round1"].SetArg(3,I2)
	this.Kernels["round1"].SetArg(4,I1)
	this.Kernels["round1"].SetArg(5,uint32(DUCK_SIZE_A*1024))
	this.Kernels["round1"].SetArg(6,uint32(DUCK_SIZE_B*1024))
	this.Kernels["roundN0"], err = this.Program.CreateKernel("FluffyRoundNO1")
	this.Kernels["roundN0"].SetArg(0,B)
	this.Kernels["roundN0"].SetArg(1,A1)
	this.Kernels["roundN0"].SetArg(2,I1)
	this.Kernels["roundN0"].SetArg(3,I2)
	this.Kernels["roundNA"], err = this.Program.CreateKernel("FluffyRoundNON")
	this.Kernels["roundNA"].SetArg(0,B)
	this.Kernels["roundNA"].SetArg(1,A1)
	this.Kernels["roundNA"].SetArg(2,I1)
	this.Kernels["roundNA"].SetArg(3,I2)
	this.Kernels["roundNB"], err = this.Program.CreateKernel("FluffyRoundNON")
	this.Kernels["roundNB"].SetArg(0,A1)
	this.Kernels["roundNB"].SetArg(1,B)
	this.Kernels["roundNB"].SetArg(2,I2)
	this.Kernels["roundNB"].SetArg(3,I1)
	this.Kernels["tail"], err = this.Program.CreateKernel("FluffyTailO")
	this.Kernels["tail"].SetArg(0,B)
	this.Kernels["tail"].SetArg(1,A1)
	this.Kernels["tail"].SetArg(2,I1)
	this.Kernels["tail"].SetArg(3,I2)
	//this.Kernels["recovery"], err = this.Program.CreateKernel("kernelRecovery")
	//this.Kernels["recovery"].SetArg(0,k[0])
	//this.Kernels["recovery"].SetArg(1,k[1])
	//this.Kernels["recovery"].SetArg(2,k[2])
	//this.Kernels["recovery"].SetArg(3,k[3])
	//this.Kernels["recovery"].SetArg(4,R)
	//this.Kernels["recovery"].SetArg(4,I2)


	//Run the kernel

	this.CommandQueue.EnqueueFillBuffer(I1,nil,0,0,0,nil)
	this.CommandQueue.EnqueueFillBuffer(I2,nil,0,0,0,nil)

	if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["seedA"], []int{0}, []int{2048*128}, []int{128}, nil); err != nil {
		log.Println("-", this.MinerId,err)
		return
	}
	if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["seedB1"], []int{0}, []int{1024*128}, []int{128}, nil); err != nil {
		log.Println("-", this.MinerId,err)
		return
	}
	if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["seedB2"], []int{0}, []int{1024*128}, []int{128}, nil); err != nil {
		log.Println("-", this.MinerId,err)
		return
	}
	this.CommandQueue.EnqueueFillBuffer(I1,nil,0,0,0,nil)

	if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["round1"], []int{0}, []int{4096*1024}, []int{1024}, nil); err != nil {
		log.Println("-", this.MinerId,err)
		return
	}
	this.CommandQueue.EnqueueFillBuffer(I2,nil,0,0,0,nil)
	if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["roundN0"], []int{0}, []int{4096*1024}, []int{1024}, nil); err != nil {
		log.Println("-", this.MinerId,err)
		return
	}
	this.CommandQueue.EnqueueFillBuffer(I1,nil,0,0,0,nil)
	if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["roundNB"], []int{0}, []int{4096*1024}, []int{1024}, nil); err != nil {
		log.Println("-", this.MinerId,err)
		return
	}
	for i:=0;i<120;i++{
		this.CommandQueue.EnqueueFillBuffer(I2,nil,0,0,0,nil)
		if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["roundNA"], []int{0}, []int{4096*1024}, []int{1024}, nil); err != nil {
			log.Println("-", this.MinerId,err)
			return
		}
		this.CommandQueue.EnqueueFillBuffer(I1,nil,0,0,0,nil)
		if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["roundNB"], []int{0}, []int{4096*1024}, []int{1024}, nil); err != nil {
			log.Println("-", this.MinerId,err)
			return
		}
	}
	this.CommandQueue.EnqueueFillBuffer(I2,nil,0,0,0,nil)

	if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["FluffyTailO"], []int{0}, []int{4096*1024}, []int{1024}, nil); err != nil {
		log.Println("-", this.MinerId,err)
		return
	}
	edges_count := make([]byte,4)
	//Get output
	if _, err = this.CommandQueue.EnqueueReadBufferByte(I2, true, 0, edges_count, nil); err != nil {
		log.Println("- ReadBuffer", this.MinerId,err)
		return
	}
	edges_count_val := binary.LittleEndian.Uint32(edges_count)
	if edges_count_val > 1000000{
		edges_count_val = 1000000
	}
	edges_left = make([]byte,edges_count_val*2)
	//Get output
	if _, err = this.CommandQueue.EnqueueReadBufferByte(A1, true, 0, edges_count, nil); err != nil {
		log.Println("- ReadBuffer", this.MinerId,err)
		return
	}

	log.Println(edges_left)
	os.Exit(1)
}

func (this *Device)Update()  {
	this.CurrentWorkID++
}

func (this *Device)InitDevice()  {

}

func (d *Device)Release()  {
	d.Context.Release()
	d.BlockObj.Release()
	d.NonceOutObj.Release()
	d.Program.Release()
	d.CommandQueue.Release()
}
