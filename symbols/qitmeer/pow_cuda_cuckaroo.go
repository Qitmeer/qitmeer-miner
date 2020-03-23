//+build cuda,!opencl

/**
Qitmeer
james
*/
package qitmeer

/*
#cgo LDFLAGS: -L../../lib/cuda
#cgo LDFLAGS: -lcudacuckoo
#cgo CFLAGS: -g -O3 -fno-stack-protector
#include "../../lib/cuckoo.h"
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/core"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"math"
	"math/big"
	"sort"
	"sync"
	"time"
	"unsafe"
)

type CudaCuckaroo struct {
	core.Device
	ClearBytes    []byte
	Work          *QitmeerWork
	header        MinerBlockData
	EdgeBits      int
	Step          int
	WorkGroupSize int
	LocalSize     int
	Nedge         int
	Edgemask      uint64
	Nonces        []uint32
	solverCtx     unsafe.Pointer
	average       [10]float64
	averageJ      int
	lock          sync.Mutex
}

func (this *CudaCuckaroo) InitDevice() {
	this.EdgeBits = this.Cfg.OptionConfig.EdgeBits
	common.MinerLoger.Debug(fmt.Sprintf("==============Mining Cuckaroo with CUDA: deviceID:%d edge bits:%d ============== module=miner", this.MinerId, this.EdgeBits))
	this.average = [10]float64{0, 0, 0, 0, 0, 0}
	this.averageJ = 1
	C.init_solver((C.int)(this.MinerId), &this.solverCtx,
		(C.int)(this.Cfg.OptionConfig.Expand),
		(C.int)(this.Cfg.OptionConfig.Ntrims),
		(C.int)(this.Cfg.OptionConfig.Genablocks),
		(C.int)(this.Cfg.OptionConfig.Genatpb),
		(C.int)(this.Cfg.OptionConfig.Genbtpb),
		(C.int)(this.Cfg.OptionConfig.Trimtpb),
		(C.int)(this.Cfg.OptionConfig.Tailtpb),
		(C.int)(this.Cfg.OptionConfig.Recoverblocks),
		(C.int)(this.Cfg.OptionConfig.Recovertpb),
	)
}

func (this *CudaCuckaroo) Update() {
	randStr := fmt.Sprintf("%s%d%d", this.Cfg.SoloConfig.RandStr, this.MinerId, this.CurrentWorkID)
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		h := md5.Sum([]byte(randStr))
		this.Work.PoolWork.ExtraNonce2 = hex.EncodeToString(h[:])[:8]
		this.header.Exnonce2 = this.Work.PoolWork.ExtraNonce2
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepQitmeerWork()
		this.header.PackagePoolHeader(this.Work, pow.CUCKAROO)
		// common.MinerLoger.Debug(fmt.Sprintf(" # %d",this.MinerId)+"ex2:" + this.header.Exnonce2+" tx root:"+this.header.HeaderBlock.TxRoot.String())
	} else {
		txHash, txs := this.Work.Block.CalcCoinBase(this.Cfg, randStr, this.CurrentWorkID, this.Cfg.SoloConfig.MinerAddr)
		this.header.PackageRpcHeader(this.Work, txs)
		this.header.HeaderBlock.TxRoot = *txHash
	}
}

func (this *CudaCuckaroo) Mine(wg *sync.WaitGroup) {
	this.AverageHashRate = 0
	defer this.Release()
	defer wg.Done()
	this.HasNewWork = false
	this.CurrentWorkID = 0
	this.header = MinerBlockData{
		Transactions: []Transactions{},
		Parents:      []ParentItems{},
		HeaderData:   make([]byte, 0),
		TargetDiff:   &big.Int{},
		JobID:        "",
	}
	this.Started = time.Now().Unix()
	wg1 := sync.WaitGroup{}
	wg1.Add(1)
	c := make(chan interface{})
	go func() {
		defer close(c)
		wg1.Wait()
	}()
	work := make(chan core.BaseWork, 1)
	isFirst := true
	go func() {
		defer wg1.Done()
		for {
			// if has new work ,current calc stop
			select {
			case w := <-work:
				this.HasNewWork = false
				for {
					if this.HasNewWork || this.ForceStop {
						break
					}
					this.Work = w.(*QitmeerWork)
					this.Update()
					if this.header.Height != common.CurrentHeight {
						common.MinerLoger.Debug("job expired!", "cheight", this.header.Height, "newheoght", common.CurrentHeight)
						break
					}
					if this.Pool && this.Work.PoolWork.JobID != common.JobID {
						common.MinerLoger.Debug("job expired!", "jobID", this.Work.PoolWork.JobID, "newJob", common.JobID)
						break
					}
					this.CardRun()
					this.IsRunning = false
					if !this.Pool {
						break
					}
				}
			}
		}
	}()

	for {
		select {
		case w := <-this.NewWork:
			if isFirst {
				work <- w
				isFirst = false
				continue
			}
			cwork := w.(*QitmeerWork)
			if this.ForceStop {
				if this.IsRunning && this.solverCtx != nil {
					C.stop_solver(this.solverCtx)
				}
				common.MinerLoger.Debug("=============force stop because network============")
				continue
			}
			if this.Pool && cwork.PoolWork.JobID != common.JobID {
				continue
			}
			if !this.Pool && cwork.Block.Height != common.CurrentHeight {
				continue
			}
			if this.IsRunning && this.solverCtx != nil {
				C.stop_solver(this.solverCtx)
			}
			work <- w
		case <-this.Quit:
			return
		case <-c:
			return
		}
		if !this.IsValid {
			return
		}
	}
}
func (this *CudaCuckaroo) CardRun() bool {
	this.Nonces = make([]uint32, 0)
	hData := this.header.HeaderBlock.BlockData()[:types.MaxBlockHeaderPayload-pow.PROOF_DATA_CIRCLE_NONCE_END]
	powStruct := this.header.HeaderBlock.Pow.(*pow.Cuckaroo)
	cycleNoncesBytes := make([]byte, 42*4*10) //max 10 answers
	nonceBytes := make([]byte, 4)
	this.average[0] = 0
	graphWeight := CuckarooGraphWeight(int64(this.header.Height), int64(this.Cfg.OptionConfig.BigGraphStartHeight), uint(this.EdgeBits))
	target := pow.CuckooDiffToTarget(graphWeight, this.header.TargetDiff)
	targetBytes, _ := hex.DecodeString(target)

	common.MinerLoger.Debug(fmt.Sprintf("========================== # %d card begin work height:%d of %d===================", this.MinerId, this.header.Height, common.CurrentHeight))
	var wg = new(sync.WaitGroup)
	c := make(chan interface{})
	wg.Add(1)
	go func() {
		defer close(c)
		wg.Wait()
	}()
	go func() {
		defer wg.Done()
		isFind := C.run_solver((C.int)(this.MinerId), this.solverCtx, (*C.char)(unsafe.Pointer(&hData[0])), (C.int)(len(hData)), 0, math.MaxUint32, (*C.uchar)(unsafe.Pointer(&targetBytes[0])),
			(*C.uint)(unsafe.Pointer(&nonceBytes[0])),
			(*C.uint)(unsafe.Pointer(&cycleNoncesBytes[0])), (*C.double)(unsafe.Pointer(&this.average[0])))
		if isFind != 1 {
			c <- "not found"
			return
		}
		// common.MinerLoger.Debug(fmt.Sprintf("# %d",this.MinerId) + "==================== Current PoolWork:================= current jobID"+this.Work.PoolWork.JobID+" CB1:"+
		// 	this.Work.PoolWork.CB1+" CB2:"+this.Work.PoolWork.CB2+ "CB3:" + this.Work.PoolWork.CB3+"CB4:" + this.Work.PoolWork.CB4 + " ntime :" + this.Work.PoolWork.Ntime)
		common.MinerLoger.Debug(fmt.Sprintf("# %d", this.MinerId) + "will submit header info :" + this.header.JobID + "-" + this.header.Exnonce2 + "-" + this.Work.PoolWork.ExtraNonce1)
		//nonce
		copy(hData[108:112], nonceBytes)
		index := 0
		for {
			if index > 9 {
				break
			}
			cbytes := cycleNoncesBytes[index*42*4 : (index+1)*42*4]
			for jj := 0; jj < len(cbytes); jj += 4 {
				tj := binary.LittleEndian.Uint32(cbytes[jj : jj+4])
				if tj <= 0 {
					isFind = 0
					break
				}
				this.Nonces = append(this.Nonces, tj)
			}
			if isFind != 1 {
				c <- "not found"
				return
			}
			sort.Slice(this.Nonces, func(i, j int) bool {
				return this.Nonces[i] < this.Nonces[j]
			})
			powStruct.SetCircleEdges(this.Nonces)
			powStruct.SetNonce(binary.LittleEndian.Uint32(nonceBytes))
			powStruct.SetEdgeBits(uint8(this.EdgeBits))
			subData := BlockDataWithProof(this.header.HeaderBlock)
			copy(subData[:113], hData[:113])
			h := hash.DoubleHashH(subData)
			common.MinerLoger.Debug(fmt.Sprintf("# %d Calc Hash %s  target diff:%d  target:%s", this.MinerId, h, this.header.TargetDiff.Uint64(), target))

			subm := hex.EncodeToString(subData)

			if !this.Pool {
				subm += common.Int2varinthex(int64(len(this.header.Parents)))
				for j := 0; j < len(this.header.Parents); j++ {
					subm += this.header.Parents[j].Data
				}
				txCount := len(this.header.Transactions)
				subm += common.Int2varinthex(int64(txCount))

				for j := 0; j < txCount; j++ {
					subm += this.header.Transactions[j].Data
				}
				subm += "-" + fmt.Sprintf("%d", txCount) + "-" + fmt.Sprintf("%d", this.header.Height)
			} else {
				subm += "-" + this.header.JobID + "-" + this.header.Exnonce2
			}
			common.MinerLoger.Debug(fmt.Sprintf("# %d", this.MinerId)+"subm:", subm)
			common.MinerLoger.Debug(fmt.Sprintf("# %d submit header job info :", this.MinerId) + this.header.JobID + "-" + this.header.Exnonce2)
			this.SubmitData <- subm
			index++
		}
		c <- nil
	}()
	this.IsRunning = true
	for {
		select {
		case err := <-c:
			if err == nil {
				return true
			}
			return false
		}
	}
}
func (this *CudaCuckaroo) Status(wg *sync.WaitGroup) {
	defer wg.Done()
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for {
		select {
		case <-this.Quit:
			return
		case <-t.C:

			if !this.IsValid {
				return
			}
		calcAverageHash:
			for i := 0; i < 10; i++ {
				if this.average[i] <= 0 {
					continue
				}
				count := 0
				for j := 0; j < 10; j++ {
					if i == j || this.average[j] <= 0 {
						continue
					}
					if math.Abs(float64(this.average[i]-this.average[j])) < 2 {
						count++
					}
					if count >= 4 {
						this.AverageHashRate = this.average[i]
						break calcAverageHash
					}
					if j > 6 {
						break
					}
				}
			}
			if this.AverageHashRate > 0 {
				common.MinerLoger.Info(fmt.Sprintf("# %d [%s] : %f GPS", this.MinerId, this.ClDevice.Name(), this.AverageHashRate))
			}

		}
	}
}

func (this *CudaCuckaroo) SubmitShare(substr chan string) {
	if !this.GetIsValid() {
		return
	}
	for {
		select {
		case <-this.Quit:
			return
		case str := <-this.SubmitData:
			substr <- str
		}
	}
}
