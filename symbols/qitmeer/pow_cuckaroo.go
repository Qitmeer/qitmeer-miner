/**
Qitmeer
james
*/
package qitmeer

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/go-opencl/cl"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	cuckaroo "github.com/Qitmeer/qitmeer/crypto/cuckoo"
	"github.com/Qitmeer/qitmeer/crypto/cuckoo/siphash"
	"github.com/Qitmeer/qitmeer/params"
	"math/big"
	`os`
	"qitmeer-miner/common"
	"qitmeer-miner/core"
	"qitmeer-miner/cuckoo"
	"qitmeer-miner/kernel"
	"sort"
	`strings`
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Cuckaroo struct {
	core.Device
	ClearBytes	[]byte
	EdgesObj              *cl.MemObject
	EdgesBytes            []byte
	DestinationEdgesCountObj              *cl.MemObject
	DestinationEdgesCountBytes            []byte
	EdgesIndexObj         *cl.MemObject
	EdgesIndex1Obj         *cl.MemObject
	EdgesIndexBytes       []byte
	DestinationEdgesObj   *cl.MemObject
	DestinationEdgesBytes []byte
	NoncesObj             *cl.MemObject
	NoncesBytes           []byte
	Nonces           []uint32
	NodesObj              *cl.MemObject
	NodesBytes            []byte
	Edges                 []uint32
	CreateEdgeKernel      *cl.Kernel
	Trimmer01Kernel       *cl.Kernel
	Trimmer02Kernel       *cl.Kernel
	RecoveryKernel        *cl.Kernel
	Work                  *QitmeerWork
	header MinerBlockData
	EdgeBits            int
	Step            int
	WorkGroupSize            int
	LocalSize            int
	Nedge            int
	Edgemask            uint64
}

func (this *Cuckaroo) InitDevice() {
	err := os.Setenv("GPU_MAX_HEAP_SIZE", "100")
	if err != nil {
		common.MinerLoger.Errorf(err.Error())
		this.IsValid = false
		return
	}
	err = os.Setenv("GPU_USE_SYNC_OBJECTS", "1")
	if err != nil {
		common.MinerLoger.Errorf(err.Error())
		this.IsValid = false
		return
	}
	err = os.Setenv("GPU_MAX_ALLOC_PERCENT", "100")
	if err != nil {
		common.MinerLoger.Errorf(err.Error())
		this.IsValid = false
		return
	}
	err = os.Setenv("GPU_SINGLE_ALLOC_PERCENT", "100")
	if err != nil {
		common.MinerLoger.Errorf(err.Error())
		this.IsValid = false
		return
	}
	err = os.Setenv("GPU_64BIT_ATOMICS", "100")
	if err != nil {
		common.MinerLoger.Errorf(err.Error())
		this.IsValid = false
		return
	}
	err = os.Setenv("GPU_FORCE_64BIT_PTR", "100")
	if err != nil {
		common.MinerLoger.Errorf(err.Error())
		this.IsValid = false
		return
	}
	err = os.Setenv("GPU_MAX_WORKGROUP_SIZE", "1024")
	if err != nil {
		common.MinerLoger.Errorf(err.Error())
		this.IsValid = false
		return
	}
	err = os.Setenv("CL_LOG_ERRORS", "stdout")
	if err != nil {
		common.MinerLoger.Errorf(err.Error())
		this.IsValid = false
		return
	}
	this.Device.InitDevice()
	if !this.IsValid {
		return
	}
	this.EdgeBits = this.Cfg.OptionConfig.EdgeBits

	this.Nedge = 1 << uint(this.EdgeBits)
	this.Edgemask = uint64(this.Nedge - 1)
	this.Step = 1 << (uint(this.EdgeBits)-20)
	this.WorkGroupSize = this.Cfg.OptionConfig.GroupSize
	this.LocalSize = this.Cfg.OptionConfig.LocalSize
	common.MinerLoger.Debugf("mining edge bits:%d local size:%d work size:%d",this.EdgeBits,this.LocalSize,this.WorkGroupSize)
	kernelStr := strings.ReplaceAll(kernel.CuckarooKernelNew,"{{edge_bits}}",fmt.Sprintf("%d",this.EdgeBits))
	kernelStr = strings.Replace(kernelStr,"{{step}}",fmt.Sprintf("%d",this.Step),1)
	kernelStr = strings.ReplaceAll(kernelStr,"{{group}}",fmt.Sprintf("%d",this.WorkGroupSize))
	this.Program, err = this.Context.CreateProgramWithSource([]string{kernelStr})
	if err != nil {
		common.MinerLoger.Errorf("-%d,%s%v", this.MinerId, this.DeviceName, err)
		this.IsValid = false
		return
	}

	err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
	if err != nil {
		common.MinerLoger.Errorf("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}

	this.InitKernelAndParam()

}

func (this *Cuckaroo) Update() {
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID)
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepQitmeerWork()
		this.header.PackagePoolHeader(this.Work,pow.CUCKAROO)
	} else {
		randStr := fmt.Sprintf("%s%d",this.Cfg.SoloConfig.RandStr,this.CurrentWorkID)
		txHash := this.Work.Block.CalcCoinBase(this.Cfg,randStr, this.CurrentWorkID, this.Cfg.SoloConfig.MinerAddr)
		this.header.PackageRpcHeader(this.Work)
		this.header.HeaderBlock.TxRoot = *txHash
	}
}

func (this *Cuckaroo) Mine(wg *sync.WaitGroup) {

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

		this.HasNewWork = false
		this.CurrentWorkID = 0
		this.header = MinerBlockData{
			Transactions:[]Transactions{},
			Parents:[]ParentItems{},
			HeaderData:make([]byte,0),
			TargetDiff:&big.Int{},
			JobID:"",
		}
		var err error
		this.Started = time.Now().Unix()
		this.AllDiffOneShares = 0
		for {
			// if has new work ,current calc stop
			if this.HasNewWork {
				break
			}
			this.Update()
			for {
				if this.HasNewWork {
					break
				}
				nonce ,_:= common.RandUint32()
				this.header.HeaderBlock.Pow.SetNonce(nonce)
				hdrkey := hash.HashH(this.header.HeaderBlock.BlockData())
				if this.Cfg.OptionConfig.CPUMiner{
					c := cuckaroo.NewCuckoo()
					var found = false
					this.Nonces,found = c.PoW(hdrkey[:])
					if !found || len(this.Nonces) != cuckaroo.ProofSize{
						continue
					}
				} else{
					this.DestinationEdgesBytes = make([]byte,0)
					sip := siphash.Newsip(hdrkey[:])
					this.InitParamData()
					err = this.CreateEdgeKernel.SetArg(0,uint64(sip.V[0]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.CreateEdgeKernel.SetArg(1,uint64(sip.V[1]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.CreateEdgeKernel.SetArg(2,uint64(sip.V[2]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.CreateEdgeKernel.SetArg(3,uint64(sip.V[3]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.Trimmer01Kernel.SetArg(0,uint64(sip.V[0]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.Trimmer01Kernel.SetArg(1,uint64(sip.V[1]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.Trimmer01Kernel.SetArg(2,uint64(sip.V[2]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.Trimmer01Kernel.SetArg(3,uint64(sip.V[3]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.Trimmer02Kernel.SetArg(0,uint64(sip.V[0]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.Trimmer02Kernel.SetArg(1,uint64(sip.V[1]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.Trimmer02Kernel.SetArg(2,uint64(sip.V[2]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					err = this.Trimmer02Kernel.SetArg(3,uint64(sip.V[3]))
					if err != nil {
						common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					// 2 ^ 24 2 ^ 11 * 2 ^ 8 * 2 * 2 ^ 4 11+8+1+4=24  12 + 8 + 4
					if this.Event, err = this.CommandQueue.EnqueueNDRangeKernel(this.CreateEdgeKernel, []int{0}, []int{this.LocalSize*this.WorkGroupSize}, []int{this.WorkGroupSize}, nil); err != nil {
						common.MinerLoger.Infof("CreateEdgeKernel-1058%d,%v", this.MinerId,err)
						return
					}
					this.Event.Release()
					for i:= 0;i<120;i++{
						if this.Event, err = this.CommandQueue.EnqueueNDRangeKernel(this.Trimmer01Kernel, []int{0}, []int{this.LocalSize*this.WorkGroupSize}, []int{this.WorkGroupSize}, nil); err != nil {
							common.MinerLoger.Infof("Trimmer01Kernel-1058%d,%v", this.MinerId,err)
							return
						}
						this.Event.Release()
						_ = this.CommandQueue.Finish()
					}
					if this.Event, err = this.CommandQueue.EnqueueNDRangeKernel(this.Trimmer02Kernel, []int{0}, []int{this.LocalSize*this.WorkGroupSize}, []int{this.WorkGroupSize}, nil); err != nil {
						common.MinerLoger.Infof("Trimmer02Kernel-1058%d,%v", this.MinerId,err)
						return
					}
					this.Event.Release()
					this.DestinationEdgesCountBytes = make([]byte,8)
					this.Event,err = this.CommandQueue.EnqueueReadBufferByte(this.DestinationEdgesCountObj,true,0,this.DestinationEdgesCountBytes,nil)
					if err != nil {
						common.MinerLoger.Infof("DestinationEdgesCountObj-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					this.Event.Release()
					count := binary.LittleEndian.Uint32(this.DestinationEdgesCountBytes[4:8])
					if count < cuckaroo.ProofSize*2 {
						continue
					}
					this.DestinationEdgesBytes = make([]byte,count*4)
					this.Event,err = this.CommandQueue.EnqueueReadBufferByte(this.DestinationEdgesObj,true,0,this.DestinationEdgesBytes,nil)
					if err != nil {
						common.MinerLoger.Infof("DestinationEdgesObj-%d,%v", this.MinerId, err)
						this.IsValid = false
						return
					}
					this.Event.Release()
					this.Edges = make([]uint32,0)
					for j:=0;j<len(this.DestinationEdgesBytes);j+=4{
						blockNonce := binary.LittleEndian.Uint32(this.DestinationEdgesBytes[j:j+4])
						u00 := siphash.SiphashPRF(&sip.V, uint64(blockNonce<<1))
						v00 := siphash.SiphashPRF(&sip.V, (uint64(blockNonce)<<1)|1)
						u := uint32(u00&this.Edgemask) << 1
						v := (uint32(v00&this.Edgemask) << 1) | 1
						this.Edges = append(this.Edges,u)
						this.Edges = append(this.Edges,v)
					}
					cg := cuckoo.CGraph{}
					cg.SetEdges(this.Edges,int(count))
					atomic.AddUint64(&this.AllDiffOneShares, 1)
					if !cg.FindSolutions(){
						continue
					}
					edges := cg.CycleEdges.GetData()
					this.Nonces = make([]uint32,0)
					sort.Slice(this.Nonces, func(i, j int) bool {
						return this.Nonces[i] < this.Nonces[j]
					})
					for j:=0;j<len(this.DestinationEdgesBytes);j+=4{
						blockNonce := binary.LittleEndian.Uint32(this.DestinationEdgesBytes[j:j+4])
						u00 := siphash.SiphashPRF(&sip.V, uint64(blockNonce<<1))
						v00 := siphash.SiphashPRF(&sip.V, (uint64(blockNonce)<<1)|1)
						u := uint32(u00&this.Edgemask) << 1
						v := (uint32(v00&this.Edgemask) << 1) | 1
						for i := 0;i < 42;i++{
							if (u== uint32(edges[i].Item1) && v == uint32(edges[i].Item2)) || (u== uint32(edges[i].Item2) && v == uint32(edges[i].Item1)) {
								this.Nonces = append(this.Nonces,blockNonce)
							}
						}
					}
					sort.Slice(this.Nonces, func(i, j int) bool {
						return this.Nonces[i] < this.Nonces[j]
					})
				}
				// when GPU find cuckoo cycle one time GPS/s
				this.AllDiffOneShares += 1
				powStruct := this.header.HeaderBlock.Pow.(*pow.Cuckaroo)
				powStruct.SetCircleEdges(this.Nonces)
				powStruct.SetEdgeBits(uint8(this.EdgeBits))
				powStruct.SetScale(uint32(params.MixNetParams.PowConfig.CuckarooDiffScale))
				powStruct.SetNonce(nonce)
				err := cuckaroo.VerifyCuckaroo(hdrkey[:],this.Nonces[:],uint(this.EdgeBits))
				if err != nil{
					common.MinerLoger.Errorf("[error]Verify Error!%v",err)
					continue
				}
				targetDiff := pow.CompactToBig(this.header.HeaderBlock.Difficulty)
				if pow.CalcCuckooDiff(int64(params.MixNetParams.PowConfig.CuckarooDiffScale),powStruct.GetBlockHash([]byte{})) < targetDiff.Uint64(){
					common.MinerLoger.Error("difficulty is too easy!")
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

func (this *Cuckaroo) SubmitShare(substr chan string) {
	this.Device.SubmitShare(substr)
}

func (this *Cuckaroo) Release() {
	this.Context.Release()
	this.Program.Release()
	this.CreateEdgeKernel.Release()
	this.Trimmer01Kernel.Release()
	this.Trimmer02Kernel.Release()
	this.EdgesObj.Release()
	this.EdgesIndexObj.Release()
	this.EdgesIndex1Obj.Release()
	this.DestinationEdgesObj.Release()
}

func (this *Cuckaroo) InitParamData() {
	var err error
	this.ClearBytes = make([]byte,4)
	this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesIndexObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,this.Nedge*4,nil)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.Event.Release()
	this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesIndex1Obj,unsafe.Pointer(&this.ClearBytes[0]),4,0,this.Nedge*4,nil)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.Event.Release()
	this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,this.Nedge*2,nil)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.Event.Release()
	this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.DestinationEdgesObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,this.Nedge*2,nil)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.Event.Release()
	this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.DestinationEdgesCountObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,8,nil)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.Event.Release()
	err = this.CreateEdgeKernel.SetArgBuffer(4,this.EdgesObj)
	err = this.CreateEdgeKernel.SetArgBuffer(5,this.EdgesIndexObj)
	err = this.CreateEdgeKernel.SetArgBuffer(6,this.EdgesIndex1Obj)

	err = this.Trimmer01Kernel.SetArgBuffer(4,this.EdgesObj)
	err = this.Trimmer01Kernel.SetArgBuffer(5,this.EdgesIndexObj)
	err = this.Trimmer01Kernel.SetArgBuffer(6,this.EdgesIndex1Obj)

	err = this.Trimmer02Kernel.SetArgBuffer(4,this.EdgesObj)
	err = this.Trimmer02Kernel.SetArgBuffer(5,this.EdgesIndexObj)
	err = this.Trimmer02Kernel.SetArgBuffer(6,this.EdgesIndex1Obj)
	err = this.Trimmer02Kernel.SetArgBuffer(7,this.DestinationEdgesObj)
	err = this.Trimmer02Kernel.SetArgBuffer(8,this.DestinationEdgesCountObj)
}

func (this *Cuckaroo) InitKernelAndParam() {
	var err error
	this.CreateEdgeKernel, err = this.Program.CreateKernel("CreateEdges")
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}

	this.Trimmer01Kernel, err = this.Program.CreateKernel("Trimmer01")
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}

	this.Trimmer02Kernel, err = this.Program.CreateKernel("Trimmer02")
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}

	this.EdgesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, this.Nedge*2)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.DestinationEdgesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, this.Nedge*2)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.EdgesIndexObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, this.Nedge*4)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.EdgesIndex1Obj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, this.Nedge*4)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.DestinationEdgesCountObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, 8)
	if err != nil {
		common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
		this.IsValid = false
		return
	}
}
