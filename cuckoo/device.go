package cuckoo

import (
	"encoding/binary"
	"fmt"
	"github.com/robvanmieghem/go-opencl/cl"
	"golang.org/x/crypto/blake2b"
	"log"
	"os"
	"sort"
	"sync"
	"time"
	"unsafe"
)
const EDGE_INDEX  = 24
const EDGE_SIZE  = 1 << EDGE_INDEX
const edgemask  = EDGE_SIZE - 1
const easiness  = 2 * EDGE_SIZE
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
}

func (this *Device)Mine()  {
	var err error
	this.Context, err = cl.CreateContext([]*cl.Device{this.ClDevice})
	if err != nil {
		log.Println("-1", this.MinerId, err)
		return
	}
	this.CommandQueue, err = this.Context.CreateCommandQueue(this.ClDevice, 0)
	if err != nil {
		log.Println("-2", this.MinerId,  err)
	}
	this.Program, err = this.Context.CreateProgramWithSource([]string{NewKernel})
	if err != nil {
		log.Println("-3", this.MinerId, this.DeviceName, err)
		return
	}

	err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
	if err != nil {
		log.Println("-build", this.MinerId, err)
		return
	}

	this.Kernels = make(map[string]*cl.Kernel,0)
	this.Kernels["seedA"], err = this.Program.CreateKernel("CreateEdges")



	B, err := this.Context.CreateEmptyBuffer(cl.MemReadWrite, EDGE_SIZE*2*4)
	C1, err := this.Context.CreateEmptyBuffer(cl.MemReadWrite, EDGE_SIZE*2*4)
	D, err := this.Context.CreateEmptyBuffer(cl.MemReadWrite, PROOF_SIZE*4*2)
	//E, err := this.Context.CreateEmptyBuffer(cl.MemReadWrite, PROOF_SIZE*8*4)
	I3, err := this.Context.CreateEmptyBuffer(cl.MemReadWrite, EDGE_SIZE*4*2)
	I4, err := this.Context.CreateEmptyBuffer(cl.MemReadWrite, 8)
	I5, err := this.Context.CreateEmptyBuffer(cl.MemReadWrite, PROOF_SIZE*4)

	//err = this.Kernels["seedA"].SetArgBuffer(6,E)
	this.Kernels["seedB1"], err = this.Program.CreateKernel("Trimmer01")

	this.Kernels["seedB2"], err = this.Program.CreateKernel("Trimmer02")

	this.Kernels["seedB3"], err = this.Program.CreateKernel("RecoveryNonce")
	err = this.Kernels["seedA"].SetArgBuffer(4,B)
	err = this.Kernels["seedA"].SetArgBuffer(5,I3)


	err = this.Kernels["seedB1"].SetArgBuffer(0,B)
	err = this.Kernels["seedB1"].SetArgBuffer(1,I3)

	err = this.Kernels["seedB2"].SetArgBuffer(0,B)
	err = this.Kernels["seedB2"].SetArgBuffer(1,I3)
	err = this.Kernels["seedB2"].SetArgBuffer(2,C1)
	err = this.Kernels["seedB2"].SetArgBuffer(3,I4)

	err = this.Kernels["seedB3"].SetArgBuffer(0,B)
	err = this.Kernels["seedB3"].SetArgBuffer(1,D)
	err = this.Kernels["seedB3"].SetArgBuffer(2,I5)
	defer func() {
		this.Release()
		B.Release()
		D.Release()
		I3.Release()
		I4.Release()
		I5.Release()
		C1.Release()
	}()
	for i:=725;i<2<<32;i++{
		start := int(time.Now().Unix())
		str := fmt.Sprintf("gyugguguyoadswerdwef8762565ewg82rldtest%d",i)
		hdrkey := blake2b.Sum256([]byte(str))

		sip := Newsip(hdrkey[:])
		//init kernels
		err = this.Kernels["seedA"].SetArg(0,uint64(sip.V[0]))

		err = this.Kernels["seedA"].SetArg(1,uint64(sip.V[1]))

		err = this.Kernels["seedA"].SetArg(2,uint64(sip.V[2]))

		err = this.Kernels["seedA"].SetArg(3,uint64(sip.V[3]))


		clear := make([]byte,4)
		_,err = this.CommandQueue.EnqueueFillBuffer(I3,unsafe.Pointer(&clear[0]),4,0,EDGE_SIZE*8,nil)
		_,err = this.CommandQueue.EnqueueFillBuffer(B,unsafe.Pointer(&clear[0]),4,0,EDGE_SIZE*8,nil)
		_,err = this.CommandQueue.EnqueueFillBuffer(C1,unsafe.Pointer(&clear[0]),4,0,EDGE_SIZE*8,nil)
		_,err = this.CommandQueue.EnqueueFillBuffer(D,unsafe.Pointer(&clear[0]),4,0,PROOF_SIZE*8,nil)
		//_,err = this.CommandQueue.EnqueueFillBuffer(E,unsafe.Pointer(&clear[0]),4,0,EDGE_SIZE*8*4,nil)
		_,err = this.CommandQueue.EnqueueFillBuffer(I4,unsafe.Pointer(&clear[0]),4,0,8,nil)
		_,err = this.CommandQueue.EnqueueFillBuffer(I5,unsafe.Pointer(&clear[0]),4,0,PROOF_SIZE*4,nil)

		if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["seedA"], []int{0}, []int{2048*256*2}, []int{256}, nil); err != nil {
			log.Println("1058", this.MinerId,err)
			return
		}

		result3 := make([]byte,EDGE_SIZE*8)
		edges := make([]uint32,0)
		wg := sync.WaitGroup{}
		for i:= 0;i<80;i++{
			wg.Add(1)
			go func() {
				if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["seedB1"], []int{0}, []int{2048*256*2}, []int{256}, nil); err != nil {
					log.Println("1058", this.MinerId,err)
					return
				}
				wg.Done()
			}()
		}
		wg.Wait()
		if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["seedB2"], []int{0}, []int{2048*256*2}, []int{256}, nil); err != nil {
			log.Println("1058", this.MinerId,err)
			return
		}
		result3 = make([]byte,8)
		_,err = this.CommandQueue.EnqueueReadBufferByte(I4,true,0,result3,nil)
		count := binary.LittleEndian.Uint32(result3[4:8])
		log.Println("第",i,"个，数量:",count)
		end := int(time.Now().Unix())
		log.Println("spend ",(end-start),"s one time")
		if count >= PROOF_SIZE*2{
			result2 := make([]byte,count*4*2)
			_,err = this.CommandQueue.EnqueueReadBufferByte(C1,true,0,result2,nil)
			edges = make([]uint32,0)
			for j:=0;j<len(result2);j+=4{
				edges = append(edges,binary.LittleEndian.Uint32(result2[j:j+4]))
			}
			//log.Println(i,len(edges))
			cg := CGraph{}
			cg.SetEdges(edges,len(edges)/2)
			if cg.FindSolutions(){
			//if cg.FindCycle(){
				_,err = this.CommandQueue.EnqueueFillBuffer(I5,unsafe.Pointer(&clear[0]),4,0,PROOF_SIZE*8,nil)
				_,err = this.CommandQueue.EnqueueWriteBufferByte(D,true,0,cg.GetNonceEdgesBytes(),nil)
				if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernels["seedB3"], []int{0}, []int{2048*256*2}, []int{256}, nil); err != nil {
					log.Println("1058", this.MinerId,err)
					return
				}
				nonces := make([]byte,4*PROOF_SIZE)
				_,err = this.CommandQueue.EnqueueReadBufferByte(I5,true,0,nonces,nil)
				FoundNonce := make([]uint32,0)
				for j := 0;j<PROOF_SIZE*4;j+=4{
					FoundNonce = append(FoundNonce,binary.LittleEndian.Uint32(nonces[j:j+4]))
				}

				sort.Slice(FoundNonce, func(i, j int) bool {
					return FoundNonce[i] < FoundNonce[j]
				})
				//log.Println(len(FoundNonce),FoundNonce)
				if err = Verify(hdrkey[:],FoundNonce);err == nil{
					log.Println("【Success Found Nonce】",FoundNonce)
				} else{
					log.Println("result not match:",err)
				}
				if CheckDiff(FoundNonce){
					return
				}
			}
		}
	}
}

func (this *Device)Update()  {
	this.CurrentWorkID++
}

func (this *Device)InitDevice()  {

}

func (d *Device)Release()  {
	d.Context.Release()
	d.Program.Release()
	d.CommandQueue.Release()
	for _,v:=range d.Kernels{
		v.Release()
	}
}
