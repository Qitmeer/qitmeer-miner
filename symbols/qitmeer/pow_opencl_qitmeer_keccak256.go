/**
Qitmeer
james
*/
package qitmeer

/*
#include <stddef.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/go-opencl/cl"
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/core"
	"github.com/Qitmeer/qitmeer-miner/kernel"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
)

type OpenCLKeccak256 struct {
	core.Device
	Work   *QitmeerWork
	header MinerBlockData
}

func (this *OpenCLKeccak256) InitDevice() {
	this.Started = time.Now().Unix()
	this.Device.InitDevice()
	if !this.IsValid {
		return
	}
	this.Program, this.Err = this.Context.CreateProgramWithSource([]string{kernel.QitmeerKeccak256kernelSource})
	if this.Err != nil {
		common.MinerLoger.Error(fmt.Sprintf("#-%d,%s,%v CreateProgramWithSource", this.MinerId, this.DeviceName, this.Err))
		this.IsValid = false
		return
	}

	this.Err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
	if this.Err != nil {
		common.MinerLoger.Error(fmt.Sprintf("1111-%d,%v BuildProgram", this.MinerId, this.Err))
		this.IsValid = false
		return
	}

	this.Kernel, this.Err = this.Program.CreateKernel("search")
	if this.Err != nil {
		common.MinerLoger.Error(fmt.Sprintf("-%d,%v CreateKernel", this.MinerId, this.Err))
		this.IsValid = false
		return
	}
	this.BlockObj, this.Err = this.Context.CreateEmptyBuffer(cl.MemReadOnly, 120)
	if this.Err != nil {
		common.MinerLoger.Error(fmt.Sprintf("-%d,%v CreateEmptyBuffer BlockObj", this.MinerId, this.Err))
		this.IsValid = false
		return
	}
	_ = this.Kernel.SetArgBuffer(0, this.BlockObj)
	this.NonceOutObj, this.Err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, 4)
	if this.Err != nil {
		common.MinerLoger.Error(fmt.Sprintf("-%d,%v CreateEmptyBuffer NonceOutObj", this.MinerId, this.Err))
		this.IsValid = false
		return
	}
	_ = this.Kernel.SetArgBuffer(1, this.NonceOutObj)
	this.Target2Obj, this.Err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, 32)
	if this.Err != nil {
		common.MinerLoger.Error(fmt.Sprintf("-%d,%v CreateEmptyBuffer Target2Obj", this.MinerId, this.Err))
		this.IsValid = false
		return
	}
	_ = this.Kernel.SetArgBuffer(2, this.Target2Obj)
	this.LocalItemSize = this.Cfg.OptionConfig.WorkSize
	common.MinerLoger.Debug(fmt.Sprintf("==============Mining OpenCLKeccak256=============="))
	common.MinerLoger.Debug(fmt.Sprintf("- Device ID:%d- Global item size:%d- Local item size:%d", this.MinerId, this.GlobalItemSize, this.LocalItemSize))
	this.NonceOut = make([]byte, 4)
	if this.Event, this.Err = this.CommandQueue.EnqueueWriteBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); this.Err != nil {
		common.MinerLoger.Error(fmt.Sprintf("-%d %v EnqueueWriteBufferByte NonceOutObj", this.MinerId, this.Err))
		this.IsValid = false
		return
	}
	this.Event.Release()
}

func (this *OpenCLKeccak256) Update() {
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID<<this.MinerId)[:8]
		this.header.Exnonce2 = this.Work.PoolWork.ExtraNonce2
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepQitmeerWork()
		this.header.PackagePoolHeader(this.Work, pow.QITMEERKECCAK256)
	} else {
		randStr := fmt.Sprintf("%s%d%d", this.Cfg.SoloConfig.RandStr, this.MinerId, this.CurrentWorkID)
		txHash, txs := this.Work.Block.CalcCoinBase(this.Cfg, randStr, this.CurrentWorkID, this.Cfg.SoloConfig.MinerAddr)
		this.header.PackageRpcHeader(this.Work, txs)
		this.header.HeaderBlock.TxRoot = *txHash
	}
}

func (this *OpenCLKeccak256) Mine(wg *sync.WaitGroup) {
	defer wg.Done()
	defer this.Release()
	var w core.BaseWork
	offset := 0
	for {

		select {
		case w = <-this.NewWork:
			this.Work = w.(*QitmeerWork)
		case <-this.Quit:
			return

		}
		if !this.IsValid {
			return
		}
		if this.ForceStop {
			continue
		}
		if !this.HasNewWork || this.Work == nil {
			continue
		}

		if len(this.Work.PoolWork.WorkData) <= 0 && this.Work.Block.Height <= 0 {
			continue
		}
		this.Started = time.Now().Unix()
		this.AllDiffOneShares = 0
		this.HasNewWork = false
		this.CurrentWorkID = 0
		this.header = MinerBlockData{
			Transactions: []Transactions{},
			Parents:      []ParentItems{},
			HeaderData:   make([]byte, 0),
			TargetDiff:   &big.Int{},
			JobID:        "",
		}

		for {
			// if has new work ,current calc stop
			if this.HasNewWork || this.ForceStop {
				break
			}
			this.Update()
			var err error
			hData := make([]byte, 113)
			copy(hData[0:types.MaxBlockHeaderPayload-pow.PROOFDATA_LENGTH], this.header.HeaderBlock.BlockData())
			hData = append(hData, []byte{129, 0, 0, 0, 0, 0, 0}...) //append 0x80
			if this.Event, err = this.CommandQueue.EnqueueWriteBufferByte(this.BlockObj, true, 0, hData, nil); err != nil {
				common.MinerLoger.Error(fmt.Sprintf("-%d %v EnqueueWriteBufferByte", this.MinerId, err))
				this.IsValid = false
				return
			}
			this.Event.Release()
			if !this.IsValid {
				break
			}
			if this.Event, err = this.CommandQueue.EnqueueWriteBufferByte(this.Target2Obj, true, 0, this.header.Target2, nil); err != nil {
				common.MinerLoger.Error(fmt.Sprintf("-%d %v", this.MinerId, err))
				this.IsValid = false
				return
			}
			this.Event.Release()
			if !this.IsValid {
				break
			}

			//Run the kernel
			if this.Event, this.Err = this.CommandQueue.EnqueueNDRangeKernel(this.Kernel, []int{int(offset)}, []int{this.GlobalItemSize}, []int{this.LocalItemSize}, nil); this.Err != nil {
				common.MinerLoger.Error(fmt.Sprintf("-%d %v EnqueueNDRangeKernel Kernel", this.MinerId, this.Err))
				this.IsValid = false
				return
			}
			this.Event.Release()
			this.NonceOut = make([]byte, 4)
			//Get output
			if this.Event, this.Err = this.CommandQueue.EnqueueReadBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); this.Err != nil {
				common.MinerLoger.Error(fmt.Sprintf("-%d %v EnqueueReadBufferByte NonceOutObj", this.MinerId, this.Err))
				this.IsValid = false
				return
			}
			this.Event.Release()
			atomic.AddUint64(&this.AllDiffOneShares, uint64(this.GlobalItemSize))
			xnonce := binary.LittleEndian.Uint32(this.NonceOut[0:4])
			//Found Hash
			this.header.HeaderBlock.Pow.SetNonce(xnonce)
			copy(hData[108:112], this.NonceOut)
			h := hash.HashQitmeerKeccak256(hData[:113])
			headerData := BlockDataWithProof(this.header.HeaderBlock)
			copy(headerData[0:113], hData[0:113])
			if HashToBig(&h).Cmp(this.header.TargetDiff) <= 0 {
				common.MinerLoger.Debug(fmt.Sprintf("device #%d found hash : %s nonce:%d target:%064x", this.MinerId, h, xnonce, this.header.TargetDiff))
				subm := hex.EncodeToString(headerData)
				if !this.Pool {
					subm += common.Int2varinthex(int64(len(this.header.Parents)))
					for j := 0; j < len(this.header.Parents); j++ {
						subm += this.header.Parents[j].Data
					}

					txCount := len(this.header.Transactions) //real transaction count except coinbase
					subm += common.Int2varinthex(int64(txCount))

					for j := 0; j < txCount; j++ {
						subm += this.header.Transactions[j].Data
					}
					subm += "-" + fmt.Sprintf("%d", txCount) + "-" + fmt.Sprintf("%d", this.Work.Block.Height)
				} else {
					subm += "-" + this.header.JobID + "-" + this.header.Exnonce2
				}
				this.SubmitData <- subm
				if !this.Pool {
					//solo wait new task
					this.ClearNonceData()
				}
				break
			}
			this.ClearNonceData()
		}
	}
}

func (this *OpenCLKeccak256) ClearNonceData() {
	this.NonceOut = make([]byte, 4)
	if this.Event, this.Err = this.CommandQueue.EnqueueWriteBufferByte(this.NonceOutObj, true, 0, this.NonceOut, nil); this.Err != nil {
		common.MinerLoger.Error(fmt.Sprintf("-%d %v EnqueueWriteBufferByte", this.MinerId, this.Err))
		this.IsValid = false
		return
	}
	this.Event.Release()
}
