/**
HLC FOUNDATION
james
*/
package hlc

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"qitmeer/common/hash"
	"qitmeer/core/blockchain"
	"github.com/robvanmieghem/go-opencl/cl"
	"log"
	"hlc-miner/common"
	"hlc-miner/core"
	"sync/atomic"
)

type HLCDevice struct {
	core.Device
	NewWork chan HLCWork
	Work    HLCWork
}

func (this *HLCDevice) InitDevice() {
	this.Device.InitDevice()
	if !this.IsValid {
		return
	}
	var err error
	this.Program, err = this.Context.CreateProgramWithSource([]string{kernelSource})
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
	this.Kernel.SetArgBuffer(0, this.BlockObj)
	this.NonceOutObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, 8)
	if err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
	this.Kernel.SetArgBuffer(1, this.NonceOutObj)
	this.LocalItemSize, err = this.Kernel.WorkGroupSize(this.ClDevice)
	this.LocalItemSize = this.Cfg.WorkSize
	if err != nil {
		log.Println("- WorkGroupSize failed -", this.MinerId, err)
		this.IsValid = false
		return
	}
	log.Println("- Device ID:", this.MinerId, "- Global item size:", this.GlobalItemSize, "(Intensity", this.Cfg.Intensity, ")", "- Local item size:", this.LocalItemSize)
	this.NonceOut = make([]byte, 8, 8)
	if _, err = this.CommandQueue.EnqueueWriteBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); err != nil {
		log.Println("-", this.MinerId, err)
		this.IsValid = false
		return
	}
}

func (this *HLCDevice) Update() {
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		//this.CurrentWorkID = 0
		//randstr := fmt.Sprintf("%dhlcminer%d",this.CurrentWorkID,this.MinerId)
		//byt := []byte(randstr)[:4]
		//this.Work.PoolWork.ExtraNonce2 = hex.EncodeToString(byt)
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID)
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepHlcWork()
	} else {
		randStr := fmt.Sprintf("%s%d%d", this.Cfg.RandStr, this.MinerId, this.CurrentWorkID)
		var err error
		err = this.Work.Block.CalcCoinBase(randStr, this.Cfg.MinerAddr)
		if err != nil {
			log.Println("calc coinbase error :", err)
			return
		}
		this.Work.Block.BuildMerkleTreeStore()
	}
}

func (this *HLCDevice) Mine() {

	defer this.Release()

	for {
		select {
		case this.Work = <-this.NewWork:
		case <-this.Quit:
			return

		}
		if !this.IsValid {
			continue
		}
		//if !this.Work.StartWork{
		//	continue
		//}

		if len(this.Work.PoolWork.WorkData) <= 0 && this.Work.Block.Height <= 0 {
			continue
		}

		this.HasNewWork = false
		offset := this.MinerId
		this.CurrentWorkID = 0
		for {
			// if has new work ,current calc stop
			if this.HasNewWork {
				break
			}

			this.Update()
			var header MinerBlockData
			if this.Pool {
				header.PackagePoolHeader(&this.Work)
			} else {
				header.PackageRpcHeader(&this.Work)
			}

			var err error
			if _, err = this.CommandQueue.EnqueueWriteBufferByte(this.BlockObj, true, 0, header.HeaderData, nil); err != nil {
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
			offset++
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
				for i := 0; i < 8; i++ {
					header.HeaderData[i+NONCESTART] = this.NonceOut[i]
				}
				this.Work.Block.Nonce = binary.LittleEndian.Uint64(this.NonceOut)
				h := hash.DoubleHashH(header.HeaderData)

				if blockchain.HashToBig(&h).Cmp(header.TargetDiff) <= 0 {
					log.Println("[Found Hash]",hex.EncodeToString(common.Reverse(h[:])))
					subm := hex.EncodeToString(header.HeaderData)
					if !this.Pool{
						if this.Cfg.DAG{
							subm += common.Int2varinthex(int64(len(header.Parents)))
							for j := 0; j < len(header.Parents); j++ {
								subm += header.Parents[j].Data
							}
						}
						txCount := len(header.Transactions) //real transaction count except coinbase
						subm += common.Int2varinthex(int64(txCount))

						for j := 0; j < txCount; j++ {
							subm += header.Transactions[j].Data
						}
						txCount -= 1
						subm += "-" + fmt.Sprintf("%d",txCount) + "-" + fmt.Sprintf("%d",this.Work.Block.Height)
					} else {
						subm += "-" + header.JobID + "-" + this.Work.PoolWork.ExtraNonce2
					}
					this.SubmitData <- subm
					if !this.Pool{
						//solo wait new task
						break
					}
				}
			}
			this.NonceOut = make([]byte, 8, 8)
			if _, err = this.CommandQueue.EnqueueWriteBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); err != nil {
				log.Println("-", this.MinerId, err)
				this.IsValid = false
				return
			}
		}
	}
}

func (this *HLCDevice) SubmitShare(substr chan string) {
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
