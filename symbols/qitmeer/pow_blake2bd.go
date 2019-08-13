/**
Qitmeer
james
*/
package qitmeer

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/HalalChain/go-opencl/cl"
	"github.com/HalalChain/qitmeer-lib/core/types/pow"
	"log"
	"math/big"
	"qitmeer-miner/common"
	"qitmeer-miner/core"
	"qitmeer-miner/kernel"
	"sync/atomic"
)

type Blake2bD struct {
	core.Device
	Work    *QitmeerWork
	header MinerBlockData
}

func (this *Blake2bD) InitDevice() {
	this.Device.InitDevice()
	if !this.IsValid {
		return
	}
	var err error
	this.Program, err = this.Context.CreateProgramWithSource([]string{kernel.DoubleBlake2bKernelSource})
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

	this.Kernel, err = this.Program.CreateKernel("search")
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.BlockObj, err = this.Context.CreateEmptyBuffer(cl.MemReadOnly, 128)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	_ = this.Kernel.SetArgBuffer(0, this.BlockObj)
	this.NonceOutObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, 8)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	_= this.Kernel.SetArgBuffer(1, this.NonceOutObj)
	this.LocalItemSize, err = this.Kernel.WorkGroupSize(this.ClDevice)
	this.LocalItemSize = this.Cfg.OptionConfig.WorkSize
	if err != nil {
		log.Println("- WorkGroupSize failed -", this.MinerId, err)
		this.IsValid = false
		return
	}
	log.Println("- Device ID:", this.MinerId, "- Global item size:", this.GlobalItemSize, "(Intensity", this.Cfg.OptionConfig.Intensity, ")", "- Local item size:", this.LocalItemSize)
	this.NonceOut = make([]byte, 8)
	if _, err = this.CommandQueue.EnqueueWriteBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
}

func (this *Blake2bD) Update() {
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID)
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepQitmeerWork()
		this.header.PackagePoolHeader(this.Work,pow.BLAKE2BD)
	} else {
		randStr := fmt.Sprintf("%s%d",this.Cfg.SoloConfig.RandStr,this.CurrentWorkID)
		_ = this.Work.Block.CalcCoinBase(randStr,this.Cfg.SoloConfig.MinerAddr)
		txHash := this.Work.Block.BuildMerkleTreeStore(int(this.MinerId))
		this.header.PackageRpcHeader(this.Work)
		this.header.HeaderBlock.TxRoot = txHash
	}
}

func (this *Blake2bD) Mine() {
	defer this.Release()

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
		offset := 0
		this.CurrentWorkID = 0
		this.header = MinerBlockData{
			Transactions:[]Transactions{},
			Parents:[]ParentItems{},
			HeaderData:make([]byte,0),
			TargetDiff:&big.Int{},
			JobID:"",
		}
		for {
			// if has new work ,current calc stop
			if this.HasNewWork {
				break
			}

			this.Update()
			var err error
			if _, err = this.CommandQueue.EnqueueWriteBufferByte(this.BlockObj, true, 0, this.header.HeaderBlock.BlockData(), nil); err != nil {
				log.Println("-", this.MinerId, err)
				this.IsValid = false
				break
			}
			//Run the kernel
			if _, err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernel, []int{int(offset)}, []int{this.GlobalItemSize}, []int{this.LocalItemSize}, nil); err != nil {
				log.Println("-", this.MinerId, err)
				this.IsValid = false
				break
			}
			//offset++
			//Get output
			if _, err = this.CommandQueue.EnqueueReadBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); err != nil {
				log.Println("-", this.MinerId, err)
				this.IsValid = false
				break
			}
			atomic.AddUint64(&this.AllDiffOneShares, 1)
			if this.NonceOut[0] != 0 || this.NonceOut[1] != 0 || this.NonceOut[2] != 0 || this.NonceOut[3] != 0 ||
				this.NonceOut[4] != 0 || this.NonceOut[5] != 0 || this.NonceOut[6] != 0 || this.NonceOut[7] != 0 {
				//Found Hash
				this.header.HeaderBlock.Pow.SetNonce(binary.LittleEndian.Uint64(this.NonceOut))
				h := this.header.HeaderBlock.BlockHash()

				if HashToBig(&h).Cmp(this.header.TargetDiff) <= 0 {
					log.Println("#",this.MinerId,this.DeviceName," [Found Hash]",hex.EncodeToString(common.Reverse(h[:])))
					headerData := BlockDataWithProof(this.header.HeaderBlock)
					subm := hex.EncodeToString(headerData)
					if !this.Pool{
						subm += common.Int2varinthex(int64(len(this.header.Parents)))
						for j := 0; j < len(this.header.Parents); j++ {
							subm += this.header.Parents[j].Data
						}

						txCount := len(this.header.Transactions) //real transaction count except coinbase
						subm += common.Int2varinthex(int64(txCount))

						for j := 0; j < txCount; j++ {
							subm += this.header.Transactions[j].Data
						}
						txCount -= 1
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
			this.NonceOut = make([]byte, 8)
			if _, err = this.CommandQueue.EnqueueWriteBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); err != nil {
				log.Println("-", this.MinerId, err)
				this.IsValid = false
				return
			}
		}
	}
}

func (this *Blake2bD) SubmitShare(substr chan string) {
	this.Device.SubmitShare(substr)
}

func (this *Blake2bD)Status()  {
	this.Device.Status()
}

func (this *Blake2bD)Release()  {
	this.Device.Release()
}
