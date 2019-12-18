//+build cuda,!opencl

/**
Qitmeer
james
*/
package qitmeer
/*
#cgo LDFLAGS: -L../../lib/cuda
#cgo LDFLAGS: -lcudacuckoo
#include "../../lib/cuckoo.h"
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	`encoding/binary`
	"encoding/hex"
	"fmt"
	`github.com/Qitmeer/qitmeer/common/hash`
	`github.com/Qitmeer/qitmeer/core/types`
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"math/big"
	"qitmeer-miner/common"
	"qitmeer-miner/core"
	`sort`
	"sync"
	"time"
	`unsafe`
)

type CudaCuckaroo struct {
	core.Device
	ClearBytes	[]byte
	Work                  *QitmeerWork
	header MinerBlockData
	EdgeBits            int
	Step            int
	WorkGroupSize            int
	LocalSize            int
	Nedge            int
	Edgemask            uint64
	Nonces           []uint32
	solverCtx           unsafe.Pointer
	average [1]float64
}

func (this *CudaCuckaroo) InitDevice() {
	this.EdgeBits = this.Cfg.OptionConfig.EdgeBits
	common.MinerLoger.Debug(fmt.Sprintf("==============Mining Cuckaroo with CUDA: deviceID:%d edge bits:%d ============== module=miner",this.MinerId,this.EdgeBits))
}

func (this *CudaCuckaroo) Update() {
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID<<this.MinerId)[:8]
		this.header.Exnonce2 = this.Work.PoolWork.ExtraNonce2
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepQitmeerWork()
		this.header.PackagePoolHeader(this.Work,pow.CUCKAROO)
	} else {
		randStr := fmt.Sprintf("%s%d%d",this.Cfg.SoloConfig.RandStr,this.MinerId,this.CurrentWorkID)
		txHash ,txs:= this.Work.Block.CalcCoinBase(this.Cfg,randStr, this.CurrentWorkID, this.Cfg.SoloConfig.MinerAddr)
		this.header.PackageRpcHeader(this.Work,txs)
		this.header.HeaderBlock.TxRoot = *txHash
	}
}

func (this *CudaCuckaroo) Mine(wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			common.MinerLoger.Error("recover success.","error",err)
		}
	}()
	go this.ListenStopCuda()
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
			return
		}

		if len(this.Work.PoolWork.WorkData) <= 0 && this.Work.Block.Height <= 0 {
			common.Usleep(2*1000)
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
		this.Started = time.Now().Unix()
		this.AllDiffOneShares = 0
		for {
			// if has new work ,current calc stop
			if this.Cfg.OptionConfig.MiningSyncMode && this.HasNewWork {
				common.MinerLoger.Debug(fmt.Sprintf("========================== %d new task exit current ===================",this.MinerId))
				break
			}

			this.Update()
			this.Nonces = make([]uint32,0)
			hData := this.header.HeaderBlock.BlockData()[:types.MaxBlockHeaderPayload-pow.PROOF_DATA_CIRCLE_NONCE_END]

			powStruct := this.header.HeaderBlock.Pow.(*pow.Cuckaroo)

			cycleNoncesBytes := make([]byte,42*4)
			nonceBytes := make([]byte,4)
			resultBytes := make([]byte,4)
			this.average = [1]float64{0}
			target := pow.CuckooDiffToTarget(pow.GraphWeight(uint32(this.EdgeBits),int64(this.header.Height),this.Cfg.NecessaryConfig.Param.PowConfig.BigGraphStartHeight,pow.CUCKAROO),this.header.TargetDiff)
			targetBytes,_ := hex.DecodeString(target)
			common.MinerLoger.Debug(fmt.Sprintf("========================== # %d card begin work ===================",this.MinerId))
			_ = C.cuda_search((C.int)(this.MinerId),(*C.uchar)(unsafe.Pointer(&hData[0])),(*C.uint)(unsafe.Pointer(&resultBytes[0])),(*C.uint)(unsafe.Pointer(&nonceBytes[0])),
				(*C.uint)(unsafe.Pointer(&cycleNoncesBytes[0])),(*C.double)(unsafe.Pointer(&this.average[0])),&this.solverCtx,(*C.uchar)(unsafe.Pointer(&targetBytes[0])))
			this.AverageHashRate = this.average[0]
			isFind := binary.LittleEndian.Uint32(resultBytes)
			this.average[0] = 0
			if isFind != 1 {
				break
			}

			//nonce
			copy(hData[108:112],nonceBytes)
			for jj := 0;jj < len(cycleNoncesBytes);jj+=4{
				tj := binary.LittleEndian.Uint32(cycleNoncesBytes[jj:jj+4])
				if tj <=0 {
					isFind = 0
					break
				}
				this.Nonces = append(this.Nonces,tj)
			}

			if isFind != 1{
				break
			}
			sort.Slice(this.Nonces, func(i, j int) bool {
				return this.Nonces[i]<this.Nonces[j]
			})
			powStruct.SetCircleEdges(this.Nonces)
			powStruct.SetNonce(binary.LittleEndian.Uint32(nonceBytes))
			powStruct.SetEdgeBits(uint8(this.EdgeBits))
			subData := BlockDataWithProof(this.header.HeaderBlock)
			copy(subData[:113],hData[:113])
			h := hash.DoubleHashH(subData)
			common.MinerLoger.Debug(fmt.Sprintf("# %d Calc Hash %s  target diff:%d  target:%s",this.MinerId,h,this.header.TargetDiff.Uint64(),target))

			subm := hex.EncodeToString(subData)

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
				subm += "-" + fmt.Sprintf("%d",txCount) + "-" + fmt.Sprintf("%d",this.Work.Block.Height)
			} else {
				subm += "-" + this.header.JobID + "-" + this.header.Exnonce2
			}
			this.SubmitData <- subm
			break
		}
	}
}

func (this *CudaCuckaroo)Status(wg *sync.WaitGroup)  {
	return
	defer wg.Done()
	t := time.NewTicker(time.Second*15)
	defer t.Stop()
	for {
		select{
		case <- this.Quit:
			return
		case <- t.C:

			if !this.IsValid{
				return
			}
			if this.AverageHashRate > 0 {
				common.MinerLoger.Info(fmt.Sprintf("# %d [%s] : %f GPS",this.MinerId,this.ClDevice.Name(),this.AverageHashRate))
			}

		}
	}
}

func (this *CudaCuckaroo)ListenStopCuda()  {
	common.MinerLoger.Debug("listen stop card")
	for{
		select {
		case <- this.StopTaskChan:
			if this.solverCtx != nil &&this.average[0] > 0{
				common.MinerLoger.Debug("================exit cuda because new task===========","this.solverCtx",this.solverCtx)
				C.stop_solver(this.solverCtx)
				this.average[0] = 0
			}
		}
	}
}