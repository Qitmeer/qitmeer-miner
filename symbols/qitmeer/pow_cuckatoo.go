/**
Qitmeer
james
*/
package qitmeer
/*
#cgo LDFLAGS: -L../../lib/cuckoo/target/release -lcuckoo
#include "../../lib/cuckoo.h"
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/HalalChain/go-opencl/cl"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/core/types/pow"
	"github.com/HalalChain/qitmeer-lib/crypto/cuckoo/siphash"
	"log"
	"math/big"
	"qitmeer-miner/common"
	"qitmeer-miner/core"
	"qitmeer-miner/kernel"
	"sort"
	"time"
	"unsafe"
)
const RES_BUFFER_SIZE = 4000000
const LOCAL_WORK_SIZE = 256
const GLOBAL_WORK_SIZE = 1024 * LOCAL_WORK_SIZE
const SetCnt = 1
const Trim = 2
const Extract = 3
const edges_bits = 29
var el_count = (1024 * 1024 * 512 / 32) << (edges_bits - 29)
var current_mode = SetCnt
var current_uorv = 0
var trims = 128 << (edges_bits - 29)
type Cuckatoo struct {
	core.Device
	ClearBytes	[]byte
	EdgesObj              *cl.MemObject
	EdgesBytes            []byte
	DestinationEdgesCountObj              *cl.MemObject
	DestinationEdgesCountBytes            []byte
	EdgesIndexBytes       []byte
	DestinationEdgesBytes []byte
	CountersObj             *cl.MemObject
	NoncesBytes           []byte
	ResultBytes           []byte
	Nonces           []uint32
	ResultObj              *cl.MemObject
	NodesBytes            []byte
	Edges                 []uint32
	CreateEdgeKernel      *cl.Kernel
	Work                  QitmeerWork
	Transactions                  map[int][]Transactions
	header MinerBlockData
}

func (this *Cuckatoo) InitDevice() {
	this.Device.InitDevice()
	if !this.IsValid {
		return
	}
	var err error
	this.Program, err = this.Context.CreateProgramWithSource([]string{kernel.CuckatooKernel})
	if err != nil {
		log.Println("-", this.MinerId, this.DeviceName, err)
		this.IsValid = false
		return
	}

	err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}

	this.InitKernelAndParam()

}

func (this *Cuckatoo) Update() {
	this.Transactions = make(map[int][]Transactions)
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID)
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepQitmeerWork()
	} else {
		randStr := fmt.Sprintf("%s%d%d", this.Cfg.SoloConfig.RandStr, this.MinerId, this.CurrentWorkID)
		log.Println("rand string:",randStr)
		err := this.Work.Block.CalcCoinBase(randStr, this.Cfg.SoloConfig.MinerAddr)
		if err != nil {
			log.Println("calc coinbase error :", err)
			return
		}
		this.Work.Block.BuildMerkleTreeStore()
	}
}

func (this *Cuckatoo) Mine() {

	defer this.Release()

	for {
		select {
		case w := <-this.NewWork:
			this.Work = *(w.(*QitmeerWork))
		case <-this.Quit:
			return

		}
		if !this.IsValid {
			continue
		}
		if len(this.Work.PoolWork.WorkData) <= 0 && this.Work.Block.Height <= 0 {
			continue
		}
		this.HasNewWork = false
		this.CurrentWorkID = 0
		var err error
		for {
			// if has new work ,current calc stop
			if this.HasNewWork {
				break
			}
			this.header = MinerBlockData{
				Transactions:[]Transactions{},
				Parents:[]ParentItems{},
				HeaderData:make([]byte,0),
				TargetDiff:&big.Int{},
				JobID:"",
			}
			this.Update()
			for {
				if this.HasNewWork {
					break
				}
				xnonce1 := <- common.RandGenerator(2<<32)
				xnonce2 := <- common.RandGenerator(2<<32)
				//if this.Pool {
				//	this.header.PackagePoolHeaderByNonce(this.Work,uint64(xnonce))
				//} else {
				//
				//	this.header.PackageRpcHeaderByNonce(this.Work,uint64(xnonce))
				//}
				//this.Transactions[int(this.MinerId)] = make([]Transactions,0)
				//for k := 0;k<len(this.header.Transactions);k++{
				//	this.Transactions[int(this.MinerId)] = append(this.Transactions[int(this.MinerId)],Transactions{
				//		Data:this.header.Transactions[k].Data,
				//		Hash:this.header.Transactions[k].Hash,
				//		Fee:this.header.Transactions[k].Fee,
				//	})
				//}
				this.Update()
				nonce := uint64(xnonce1) + uint64(xnonce2)
				this.Work.Block.Pow.SetNonce(nonce)
				hdrkey := hash.HashH(this.Work.Block.BlockData())

					sip := siphash.Newsip(hdrkey[:])

					this.InitParamData()
				err = this.CreateEdgeKernel.SetArg(0,uint64(sip.V[0]))
				if err != nil {
					log.Println("-", this.MinerId, err)
					this.IsValid = false
					return
				}
				err = this.CreateEdgeKernel.SetArg(1,uint64(sip.V[1]))
				if err != nil {
					log.Println("-", this.MinerId, err)
					this.IsValid = false
					return
				}
				err = this.CreateEdgeKernel.SetArg(2,uint64(sip.V[2]))
				if err != nil {
					log.Println("-", this.MinerId, err)
					this.IsValid = false
					return
				}
				err = this.CreateEdgeKernel.SetArg(3,uint64(sip.V[3]))
				if err != nil {
					log.Println("-", this.MinerId, err)
					this.IsValid = false
					return
				}
				for l:=uint32(0) ;l<uint32(trims);l++{
					current_uorv = int(l & 1)
					current_mode = SetCnt
					err = this.CreateEdgeKernel.SetArg(7,uint32(current_mode))
					err = this.CreateEdgeKernel.SetArg(8,uint32(current_uorv))
					this.Enq(8)
					current_mode = Trim
					if int(l) == (trims - 1) {
						current_mode = Extract
					}
					err = this.CreateEdgeKernel.SetArg(7,uint32(current_mode))
					this.Enq(8)
					_,err = this.CommandQueue.EnqueueFillBuffer(this.CountersObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,el_count*4,nil)

				}
				this.ResultBytes = make([]byte,RES_BUFFER_SIZE*4)
				_,_ = this.CommandQueue.EnqueueReadBufferByte(this.ResultObj,true,0,this.ResultBytes,nil)
				leftEdges := binary.LittleEndian.Uint32(this.ResultBytes[4:8])
				log.Println(fmt.Sprintf("Trimmed to %d edges",leftEdges))
				noncesBytes := make([]byte,42*4)
				if common.Timeout(10*time.Second, func() {
					p := C.malloc(C.size_t(len(this.ResultBytes)))
					// copy the data into the buffer, by converting it to a Go array
					cBuf := (*[1 << 30]byte)(p)
					copy(cBuf[:], this.ResultBytes)
					C.search_circle((*C.uint)(p),(C.ulong)(C.size_t(len(this.ResultBytes))),(*C.uint)(unsafe.Pointer(&noncesBytes[0])))
					C.free(p)
				}){
					//timeout
					log.Println("timeout 重新计算",nonce)
					continue
				}
				this.Nonces = make([]uint32,0)
				isFind := true
				for jj := 0;jj < len(noncesBytes);jj+=4{
					tj := binary.LittleEndian.Uint32(noncesBytes[jj:jj+4])
					if tj <=0 {
						isFind = false
						break
					}
					this.Nonces = append(this.Nonces,tj)
				}
				if !isFind{
					log.Println("重新计算",nonce)
					continue
				}
				sort.Slice(this.Nonces, func(i, j int) bool {
					return this.Nonces[i]<this.Nonces[j]
				})
				log.Println("find",nonce)
				powStruct := this.Work.Block.Pow.(*pow.Cuckatoo)
				powStruct.SetCircleEdges(this.Nonces)
				powStruct.SetNonce(nonce)
				err := powStruct.Verify(this.Work.Block.BlockData(),uint64(this.Work.Block.Difficulty))
				if err != nil{
					log.Println("[error]",err)
					continue
				}
				log.Println("[Found Hash]",powStruct.GetBlockHash(this.Work.Block.BlockData()))
				subm := hex.EncodeToString(this.Work.Block.BlockDataWithProof())
				if !this.Pool{
					subm += common.Int2varinthex(int64(len(this.Work.Block.Parents)))
					for j := 0; j < len(this.Work.Block.Parents); j++ {
						subm += this.Work.Block.Parents[j].Data
					}

					txCount := len(this.Work.Block.Transactions)
					subm += common.Int2varinthex(int64(txCount))

					for j := 0; j < txCount; j++ {
						subm += this.Work.Block.Transactions[j].Data
					}
					txCount -= 1 //real transaction count except coinbase
					subm += "-" + fmt.Sprintf("%d",txCount) + "-" + fmt.Sprintf("%d",this.Work.Block.Height)
					log.Println("submit string:",subm)
				} else {
					subm += "-" + this.header.JobID + "-" + this.Work.PoolWork.ExtraNonce2
				}
				this.SubmitData <- subm
				if !this.Pool{
					//solo wait new task
					break
				}

			}

		}
	}
}

func (this *Cuckatoo) SubmitShare(substr chan string) {
	this.Device.SubmitShare(substr)
}

func (this *Cuckatoo) Release() {
	this.Context.Release()
	this.Program.Release()
	this.CreateEdgeKernel.Release()
	this.EdgesObj.Release()
	this.CountersObj.Release()
	this.ResultObj.Release()
}

func (this *Cuckatoo) InitParamData() {
	var err error
	this.ClearBytes = make([]byte,4)
	allBytes := []byte{255,255,255,255}
	_,err = this.CommandQueue.EnqueueFillBuffer(this.CountersObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,el_count*4,nil)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	_,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesObj,unsafe.Pointer(&allBytes[0]),4,0,el_count*4*8,nil)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	_,err = this.CommandQueue.EnqueueFillBuffer(this.ResultObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,RES_BUFFER_SIZE*4,nil)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}

	err = this.CreateEdgeKernel.SetArgBuffer(4,this.EdgesObj)
	err = this.CreateEdgeKernel.SetArgBuffer(5,this.CountersObj)
	err = this.CreateEdgeKernel.SetArgBuffer(6,this.ResultObj)
	err = this.CreateEdgeKernel.SetArg(7,uint32(current_mode))
	err = this.CreateEdgeKernel.SetArg(8,uint32(current_uorv))

	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
}

func (this *Cuckatoo) InitKernelAndParam() {
	var err error
	this.CreateEdgeKernel, err = this.Program.CreateKernel("LeanRound")
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}

	this.EdgesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, el_count*4*8)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.CountersObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, el_count*4)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.ResultObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, RES_BUFFER_SIZE*4)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
}


func (this *Cuckatoo)Status()  {
	this.Device.Status()
}

func (this *Cuckatoo) Enq(num int) {
	offset := 0
	for j:=0;j<num;j++{
		offset = j * GLOBAL_WORK_SIZE
		//log.Println(j,offset)
		// 2 ^ 24 2 ^ 11 * 2 ^ 8 * 2 * 2 ^ 4 11+8+1+4=24
		if _, err := this.CommandQueue.EnqueueNDRangeKernel(this.CreateEdgeKernel, []int{offset}, []int{GLOBAL_WORK_SIZE}, []int{LOCAL_WORK_SIZE}, nil); err != nil {
			log.Println("CreateEdgeKernel-1058", this.MinerId,err)
			return
		}
	}
}
