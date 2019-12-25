/**+build cuda,!opencl**/

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
	`encoding/hex`
	"fmt"
	`github.com/Qitmeer/qitmeer/common/hash`
	`github.com/Qitmeer/qitmeer/core/types`
	"github.com/Qitmeer/qitmeer/core/types/pow"
	`math`
	`math/big`
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
	average [10]float64
	averageJ int
}

func (this *CudaCuckaroo) InitDevice() {
	this.EdgeBits = this.Cfg.OptionConfig.EdgeBits
	common.MinerLoger.Debug(fmt.Sprintf("==============Mining Cuckaroo with CUDA: deviceID:%d edge bits:%d ============== module=miner",this.MinerId,this.EdgeBits))
	this.average = [10]float64{0,0,0,0,0,0}
	this.averageJ = 1
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
	this.AverageHashRate = 0
	defer this.Release()
	defer wg.Done()
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
	wg1 := sync.WaitGroup{}
	wg1.Add(1)
	c := make(chan interface{})
	go func() {
		defer close(c)
		wg1.Wait()
	}()
	work := make(chan core.BaseWork,1)

	go func() {
		defer wg1.Done()
		for {
			// if has new work ,current calc stop
			select {
			case w := <- work:
				this.Work = w.(*QitmeerWork)
				this.Update()
				this.CardRun()
				this.IsRunning = false
			}
		}
	}()

	for {
		select {
		case w := <-this.NewWork:
			if this.GetIsRunning(){
				this.StopTaskChan <- true
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
func (this *CudaCuckaroo)CardRun() bool{
	this.Nonces = make([]uint32,0)
	hData := this.header.HeaderBlock.BlockData()[:types.MaxBlockHeaderPayload-pow.PROOF_DATA_CIRCLE_NONCE_END]
	powStruct := this.header.HeaderBlock.Pow.(*pow.Cuckaroo)
	cycleNoncesBytes := make([]byte,42*4)
	nonceBytes := make([]byte,4)
	resultBytes := make([]byte,4)
	this.average[0] = 0
	graphWeight := CuckarooGraphWeight(int64(this.header.Height),int64(this.Cfg.OptionConfig.BigGraphStartHeight),uint(this.EdgeBits))
	target := pow.CuckooDiffToTarget(graphWeight,this.header.TargetDiff)
	targetBytes,_ := hex.DecodeString(target)
	common.MinerLoger.Debug(fmt.Sprintf("========================== # %d card begin work height:%d===================",this.MinerId,this.header.Height))
	var wg= new(sync.WaitGroup)
	c := make(chan interface{})
	wg.Add(1)
	go func() {
		defer close(c)
		wg.Wait()
	}()
	go func() {
		defer wg.Done()
		_ = C.cuda_search((C.int)(this.MinerId),(*C.uchar)(unsafe.Pointer(&hData[0])),(*C.uint)(unsafe.Pointer(&resultBytes[0])),(*C.uint)(unsafe.Pointer(&nonceBytes[0])),
			(*C.uint)(unsafe.Pointer(&cycleNoncesBytes[0])),(*C.double)(unsafe.Pointer(&this.average[0])),&this.solverCtx,(*C.uchar)(unsafe.Pointer(&targetBytes[0])))

		isFind := binary.LittleEndian.Uint32(resultBytes)
		this.average[0] = 0
		if isFind != 1 {
			c <- "not found"
			return
		}

		//nonce
		copy(hData[108:112],nonceBytes)
		for jj := 0;jj < len(cycleNoncesBytes);jj+=4{
			tj := binary.LittleEndian.Uint32(cycleNoncesBytes[jj:jj+4])
			if tj <=0 {
				isFind = 0
				c <- "not found"
				return
			}
			this.Nonces = append(this.Nonces,tj)
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
			subm += "-" + fmt.Sprintf("%d",txCount) + "-" + fmt.Sprintf("%d",this.header.Height)
		} else {
			subm += "-" + this.header.JobID + "-" + this.header.Exnonce2
		}
		this.SubmitData <- subm
		c <- nil
	}()
	this.IsRunning = true
	for{
		select {
		case err := <-c:
			if err == nil{
				return true
			}
			return false
		case <- this.StopTaskChan:
			if this.solverCtx != nil{
				C.stop_solver(this.solverCtx)
				this.average[0] = 0
			}
		}
	}
}
func (this *CudaCuckaroo)Status(wg *sync.WaitGroup)  {
	defer wg.Done()
	t := time.NewTicker(time.Second*10)
	defer t.Stop()
	for {
		select{
		case <- this.Quit:
			return
		case <- t.C:

			if !this.IsValid{
				return
			}
calcAverageHash:
			for i := 0;i<10;i++{
				if this.average[i] <= 0 {
					continue
				}
				count := 0
				for j :=0;j<10;j++{
					if i == j || this.average[j] <= 0 {
						continue
					}
					if math.Abs(float64(this.average[i] - this.average[j]))<0.5{
						count++
					}
					if count >= 6{
						this.AverageHashRate = this.average[i]
						break calcAverageHash
					}
					if j > 6{
						break
					}
				}
			}
			if this.AverageHashRate > 0 {
				if this.AverageHashRate < 1{
					fmt.Println(this.average)
				}
				common.MinerLoger.Info(fmt.Sprintf("# %d [%s] : %f GPS",this.MinerId,this.ClDevice.Name(),this.AverageHashRate))
			}

		}
	}
}

func (this *CudaCuckaroo) SubmitShare(substr chan string) {
	if !this.GetIsValid(){
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