/**
Qitmeer
james
*/
package qitmeer

import "C"
import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/core"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"math/big"
	"sync"
	"time"
)

type QitmeerKeccak256 struct {
	core.Device
	Work   *QitmeerWork
	header MinerBlockData
}

func (this *QitmeerKeccak256) InitDevice() {
	this.Started = time.Now().Unix()
	common.MinerLoger.Debug(fmt.Sprintf("==============Mining X8R16=============="))
}

func (this *QitmeerKeccak256) Update() {
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

func (this *QitmeerKeccak256) Mine(wg *sync.WaitGroup) {
	defer wg.Done()
	defer this.Release()
	var w core.BaseWork
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
		nonce := uint32(0)
		for {
			// if has new work ,current calc stop
			if this.HasNewWork || this.ForceStop {
				break
			}
			this.Update()
			hData := make([]byte, 128)
			copy(hData[0:types.MaxBlockHeaderPayload-pow.PROOFDATA_LENGTH], this.header.HeaderBlock.BlockData())
			nonce++
			b := make([]byte, 4)
			binary.LittleEndian.PutUint32(b, nonce)
			copy(hData[108:112], b)
			h := hash.HashQitmeerKeccak256(hData[:113])
			common.MinerLoger.Debug(fmt.Sprintf("%064x", this.header.TargetDiff))
			if HashToBig(&h).Cmp(this.header.TargetDiff) <= 0 {
				headerData := BlockDataWithProof(this.header.HeaderBlock)
				copy(headerData[0:113], hData[0:113])
				common.MinerLoger.Debug(fmt.Sprintf("device #%d found hash : %s nonce:%d target:%064x", this.MinerId, h, nonce, this.header.TargetDiff))
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
				fmt.Println("subm", subm)
				this.SubmitData <- subm
				break
			}
		}
	}
}
