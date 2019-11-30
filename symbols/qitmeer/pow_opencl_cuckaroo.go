//+build opencl,!cuda

/**
Qitmeer
james
*/
package qitmeer
import "C"
import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	`os`
	"qitmeer-miner/common"
	"qitmeer-miner/core"
	"sync"
	"time"
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
}

func (this *CudaCuckaroo) InitDevice() {
	common.MinerLoger.Error("AMD Not Support CUDA!")
	os.Exit(1)
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
	defer this.Release()
	defer wg.Done()
}

func (this *CudaCuckaroo) SubmitShare(substr chan string) {
	this.Device.SubmitShare(substr)
}

func (this *CudaCuckaroo)Status(wg *sync.WaitGroup)  {
	defer wg.Done()
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for {
		select{
		case <- this.Quit:
			return
		case <- t.C:
			if !this.IsValid{
				time.Sleep(2*time.Second)
				continue
			}
			//diffOneShareHashesAvg := uint64(0x00000000FFFFFFFF)
			if this.AverageHashRate <= 0{
				continue
			}
			//recent stats 95% percent
			unit := " GPS"
			common.MinerLoger.Info(fmt.Sprintf("# %d [%s] : %s",this.MinerId,this.ClDevice.Name(),common.FormatHashRate(this.AverageHashRate,unit)))
		}
	}
}