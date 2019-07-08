package cuckatoo

import (
	"fmt"
	"github.com/HalalChain/go-opencl/cl"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/crypto/cuckoo/siphash"
	"hlc-miner/common"
	"hlc-miner/core"
	"hlc-miner/kernel"
	"hlc-miner/symbols/hlc"
	"log"
	"os"
	"testing"
	"unsafe"
)
const RES_BUFFER_SIZE = 4000000
const LOCAL_WORK_SIZE = 256
const GLOBAL_WORK_SIZE = 1024 * LOCAL_WORK_SIZE
const SetCnt = 1
const Trim = 1
const Extract = 1
func TestCuckatoo(t *testing.T)  {

	var typ = common.DevicesTypesForGPUMining
	clDevices := common.GetDevices(typ)
	deviceMiner := Cuckatoo{}
	for i, device := range clDevices {
		q := make(chan os.Signal)
		deviceMiner.Init(i,device,false,q,&common.GlobalConfig{})
	}
	deviceMiner.Mine()
}

type Cuckatoo struct {
	core.Device
	ClearBytes	[]byte
	EdgesObj              *cl.MemObject
	EdgesBytes            []byte
	DestinationEdgesCountObj              *cl.MemObject
	DestinationEdgesCountBytes            []byte
	EdgesIndexObj         *cl.MemObject
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
	Work                  hlc.HLCWork
	Transactions                  map[int][]hlc.Transactions
	header hlc.MinerBlockData
}
var el_count = (1024 * 1024 * 512 / 32) << 29 - 29
var res_buf = make([]byte,RES_BUFFER_SIZE)
var current_mode = SetCnt
var current_uorv = 0
var trims = 128 << 29 - 29
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
	this.Transactions = make(map[int][]hlc.Transactions)
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID)
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepHlcWork()
	} else {
		randStr := fmt.Sprintf("%s%d%d", this.Cfg.SoloConfig.RandStr, this.MinerId, this.CurrentWorkID)
		var err error
		err = this.Work.Block.CalcCoinBase(randStr, this.Cfg.SoloConfig.MinerAddr)
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
		var err error
		text := "helloworld"


		for {

			for nonce := 0;nonce <= 1 << 32 ;nonce++{
				if this.HasNewWork {
					break
				}
				text = fmt.Sprintf("%s%d",text,nonce)
				hdrkey := hash.DoubleHashH([]byte(text))
					sip := siphash.Newsip(hdrkey[:16])

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

					// 2 ^ 24 2 ^ 11 * 2 ^ 8 * 2 * 2 ^ 4 11+8+1+4=24
					if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.CreateEdgeKernel, []int{0}, []int{1024*256}, []int{256}, nil); err != nil {
						log.Println("CreateEdgeKernel-1058", this.MinerId,err)
						return
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
	this.Trimmer01Kernel.Release()
	this.Trimmer02Kernel.Release()
	this.RecoveryKernel.Release()
	this.EdgesObj.Release()
	this.EdgesIndexObj.Release()
	this.DestinationEdgesObj.Release()
	this.NoncesObj.Release()
	this.NodesObj.Release()
}

func (this *Cuckatoo) InitParamData() {
	var err error
	this.ClearBytes = make([]byte,4)
	_,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesIndexObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,el_count*8,nil)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	_,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,el_count*8,nil)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}

	err = this.CreateEdgeKernel.SetArgBuffer(4,this.EdgesObj)
	err = this.CreateEdgeKernel.SetArgBuffer(5,this.EdgesIndexObj)
	err = this.CreateEdgeKernel.SetArg(6,uint32(current_mode))
	err = this.CreateEdgeKernel.SetArg(7,uint32(current_uorv))

}

func (this *Cuckatoo) InitKernelAndParam() {
	var err error
	this.CreateEdgeKernel, err = this.Program.CreateKernel("LeanRound")
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}

	this.EdgesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, el_count*2*4)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.EdgesIndexObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, el_count*4)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.DestinationEdgesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, RES_BUFFER_SIZE*4)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}

}


func (this *Cuckatoo)Status()  {
	this.Device.Status()
}