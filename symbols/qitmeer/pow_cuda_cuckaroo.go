/**
Qitmeer
james
*/
package qitmeer
/*
#cgo LDFLAGS: -L../../lib/cuda
#cgo LDFLAGS: -lcudacuckoo
#include "../../lib/cuckoo.h"
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
		this.Started = time.Now().Unix()
		this.AllDiffOneShares = 0
		for {
			// if has new work ,current calc stop
			if this.HasNewWork {
				break
			}
			this.Update()

			hData := this.header.HeaderBlock.BlockData()[:types.MaxBlockHeaderPayload-pow.PROOF_DATA_CIRCLE_NONCE_END]
			s := C.CString(string(hData))
			cycleNoncesBytes := make([]byte,42*4)
			nonceBytes := make([]byte,4)
			resultBytes := make([]byte,4)
			C.cuda_search((C.int)(0),s,(*C.uint)(unsafe.Pointer(&resultBytes[0])),(*C.uint)(unsafe.Pointer(&nonceBytes[0])),(*C.uint)(unsafe.Pointer(&cycleNoncesBytes[0])))
			isFind := binary.LittleEndian.Uint32(resultBytes)
			if isFind != 1 {
				continue
			}
			//nonce
			copy(hData[104:112],nonceBytes)

			subData := BlockDataWithProof(this.header.HeaderBlock)
			copy(subData[:113],hData[:113])
			h := hash.DoubleHashH(subData)
			if pow.CalcCuckooDiff(pow.GraphWeight(uint32(this.EdgeBits)),h).Cmp(this.header.TargetDiff) < 0{
				continue
			}
			common.MinerLoger.Info(fmt.Sprintf("Found Hash %s",h))

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
		}
	}
}

func (this *CudaCuckaroo) SubmitShare(substr chan string) {
	this.Device.SubmitShare(substr)
}