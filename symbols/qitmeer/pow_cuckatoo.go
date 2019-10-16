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
	"github.com/Qitmeer/go-opencl/cl"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/crypto/cuckoo"
	"github.com/Qitmeer/qitmeer/crypto/cuckoo/siphash"
	"github.com/Qitmeer/qitmeer/params"
	"math/big"
	"qitmeer-miner/common"
	"qitmeer-miner/core"
	"qitmeer-miner/kernel"
	"sort"
	"sync"
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
	Work                  *QitmeerWork
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
		common.MinerLoger.Errorf("-", this.MinerId, this.DeviceName, err)
		this.IsValid = false
		return
	}

	err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
	if err != nil {
		common.MinerLoger.Errorf("- %d %v", this.MinerId, err)
		this.IsValid = false
		return
	}

	this.InitKernelAndParam()

}

func (this *Cuckatoo) Update() {
	this.Transactions = make(map[int][]Transactions)
	this.Device.Update()
	if this.Pool {
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID)
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepQitmeerWork()
		this.header.PackagePoolHeader(this.Work,pow.CUCKATOO)
	} else {
		randStr := fmt.Sprintf("%s%d",this.Cfg.SoloConfig.RandStr,this.CurrentWorkID)
		txHash := this.Work.Block.CalcCoinBase(this.Cfg,randStr, this.CurrentWorkID, this.Cfg.SoloConfig.MinerAddr)
		this.header.PackageRpcHeader(this.Work)
		this.header.HeaderBlock.TxRoot = *txHash
	}
}

func (this *Cuckatoo) Mine(wg *sync.WaitGroup) {

	defer this.Release()
	defer wg.Done()
	for {
		select {
		case w := <-this.NewWork:
			this.Work = w.(*QitmeerWork)
		case <-this.Quit:
			return

		}
		if !this.IsValid {
			continue
		}
		if len(this.Work.PoolWork.WorkData) <= 0 && this.Work.Block.Height <= 0 {
			continue
		}
		this.header = MinerBlockData{
			Transactions:[]Transactions{},
			Parents:[]ParentItems{},
			HeaderData:make([]byte,0),
			TargetDiff:&big.Int{},
			JobID:"",
		}
		this.HasNewWork = false
		this.CurrentWorkID = 0
		var err error
		this.Started = time.Now().Unix()
		this.AllDiffOneShares = 0
		for {
			// if has new work ,current calc stop
			if this.HasNewWork {
				break
			}
			this.Update()
			for nonce := uint32(0);nonce < ^uint32(0);nonce++{
				if this.HasNewWork {
					break
				}
				this.header.HeaderBlock.Pow.SetNonce(nonce)
				hdrkey := hash.HashH(this.header.HeaderBlock.BlockData())
				sip := siphash.Newsip(hdrkey[:])
				this.InitParamData()
				err = this.CreateEdgeKernel.SetArg(0,uint64(sip.V[0]))
				if err != nil {
					common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
					this.IsValid = false
					return
				}
				err = this.CreateEdgeKernel.SetArg(1,uint64(sip.V[1]))
				if err != nil {
					common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
					this.IsValid = false
					return
				}
				err = this.CreateEdgeKernel.SetArg(2,uint64(sip.V[2]))
				if err != nil {
					common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
					this.IsValid = false
					return
				}
				err = this.CreateEdgeKernel.SetArg(3,uint64(sip.V[3]))
				if err != nil {
					common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
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
					this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.CountersObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,el_count*4,nil)
					if err != nil {
						common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
						this.IsValid = false
						return
					}
					this.Event.Release()

				}
				this.ResultBytes = make([]byte,RES_BUFFER_SIZE*4)
				this.Event,err = this.CommandQueue.EnqueueReadBufferByte(this.ResultObj,true,0,this.ResultBytes,nil)
				if err != nil {
					common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
					this.IsValid = false
					return
				}
				this.Event.Release()
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
					this.AllDiffOneShares += 1
					continue
				}
				// when GPU find cuckoo cycle one time GPS/s
				this.AllDiffOneShares += 1
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
					continue
				}
				sort.Slice(this.Nonces, func(i, j int) bool {
					return this.Nonces[i]<this.Nonces[j]
				})
				common.MinerLoger.Errorf("find",nonce)
				powStruct := this.header.HeaderBlock.Pow.(*pow.Cuckatoo)
				powStruct.SetCircleEdges(this.Nonces)
				powStruct.SetNonce(nonce)
				powStruct.SetEdgeBits(edges_bits)
				powStruct.SetScale(uint32(params.MixNetParams.PowConfig.CuckatooDiffScale))
				err := cuckoo.VerifyCuckatoo(hdrkey[:],this.Nonces[:],uint(edges_bits))
				if err != nil{
					common.MinerLoger.Errorf("[error]Verify Error!",err)
					continue
				}
				targetDiff := pow.CompactToBig(this.header.HeaderBlock.Difficulty)
				if pow.CalcCuckooDiff(int64(params.MixNetParams.PowConfig.CuckatooDiffScale),powStruct.GetBlockHash([]byte{})) < targetDiff.Uint64(){
					common.MinerLoger.Errorf("difficulty is too easy!")
					continue
				}
				common.MinerLoger.Infof("[Found Hash]%s",this.header.HeaderBlock.BlockHash())
				subm := hex.EncodeToString(BlockDataWithProof(this.header.HeaderBlock))
				if !this.Pool{
					subm += common.Int2varinthex(int64(len(this.header.Parents)))
					for j := 0; j < len(this.header.Parents); j++ {
						subm += this.header.Parents[j].Data
					}

					txCount := len(this.header.Transactions)
					subm += common.Int2varinthex(int64(txCount))

					for j := 0; j < txCount; j++ {
						subm += this.header.Transactions[j].Data
					}
					txCount -= 1 //real transaction count except coinbase
					subm += "-" + fmt.Sprintf("%d",txCount) + "-" + fmt.Sprintf("%d",this.Work.Block.Height)
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
	this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.CountersObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,el_count*4,nil)
	if err != nil {
		common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.Event.Release()
	this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesObj,unsafe.Pointer(&allBytes[0]),4,0,el_count*4*8,nil)
	if err != nil {
		common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.Event.Release()
	this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.ResultObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,RES_BUFFER_SIZE*4,nil)
	if err != nil {
		common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.Event.Release()
	err = this.CreateEdgeKernel.SetArgBuffer(4,this.EdgesObj)
	err = this.CreateEdgeKernel.SetArgBuffer(5,this.CountersObj)
	err = this.CreateEdgeKernel.SetArgBuffer(6,this.ResultObj)
	err = this.CreateEdgeKernel.SetArg(7,uint32(current_mode))
	err = this.CreateEdgeKernel.SetArg(8,uint32(current_uorv))

	if err != nil {
		common.MinerLoger.Errorf("-", this.MinerId, err)
		this.IsValid = false
		return
	}
}

func (this *Cuckatoo) InitKernelAndParam() {
	var err error
	this.CreateEdgeKernel, err = this.Program.CreateKernel("LeanRound")
	if err != nil {
		common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
		this.IsValid = false
		return
	}

	this.EdgesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, el_count*4*8)
	if err != nil {
		common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.CountersObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, el_count*4)
	if err != nil {
		common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.ResultObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, RES_BUFFER_SIZE*4)
	if err != nil {
		common.MinerLoger.Errorf("-%d %v", this.MinerId, err)
		this.IsValid = false
		return
	}
}

func (this *Cuckatoo) Enq(num int) {
	offset := 0
	for j:=0;j<num;j++{
		offset = j * GLOBAL_WORK_SIZE
		//common.MinerLoger.Errorf(j,offset)
		// 2 ^ 24 2 ^ 11 * 2 ^ 8 * 2 * 2 ^ 4 11+8+1+4=24
		if this.Event, this.Err = this.CommandQueue.EnqueueNDRangeKernel(this.CreateEdgeKernel, []int{offset}, []int{GLOBAL_WORK_SIZE}, []int{LOCAL_WORK_SIZE}, nil); this.Err != nil {
			common.MinerLoger.Errorf("CreateEdgeKernel-1058 %d %v", this.MinerId,this.Err)
			return
		}
		this.Event.Release()
	}
}