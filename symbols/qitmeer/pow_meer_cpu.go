//+build cpu,!asic

/**
Qitmeer
james
*/
package qitmeer

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
	"strconv"
	"strings"
	"sync"
	"time"
)

type MeerCrypto struct {
	core.Device
	Work   *QitmeerWork
	header MinerBlockData
}

func (this *MeerCrypto) InitDevice() {
	this.Started = time.Now().Unix()
	common.MinerLoger.Debug(fmt.Sprintf("CPUMiner [%d] ==============Mining MeerCrypto==============", this.MinerId))
}

func (this *MeerCrypto) Update() {
	//update coinbase tx hash
	this.Device.Update()
	if this.Pool {
		this.Work.PoolWork.ExtraNonce2 = fmt.Sprintf("%08x", this.CurrentWorkID<<this.MinerId)[:8]
		this.header.Exnonce2 = this.Work.PoolWork.ExtraNonce2
		this.Work.PoolWork.WorkData = this.Work.PoolWork.PrepQitmeerWork()
		this.header.PackagePoolHeader(this.Work, pow.MEERXKECCAKV1)
	} else {
		randStr := fmt.Sprintf("%s%d%d", this.Cfg.SoloConfig.RandStr, this.MinerId, this.CurrentWorkID)
		txHash, txs := this.Work.Block.CalcCoinBase(this.Cfg, randStr, this.CurrentWorkID, this.Cfg.SoloConfig.MinerAddr)
		this.header.PackageRpcHeader(this.Work, txs)
		this.header.HeaderBlock.TxRoot = *txHash
	}
}

func (this *MeerCrypto) Mine(wg *sync.WaitGroup) {
	defer wg.Done()
	defer this.Release()
	var w core.BaseWork
	this.Started = time.Now().Unix()
	this.AllDiffOneShares = 0
	for {
		select {
		case w = <-this.NewWork:
			this.Work = w.(*QitmeerWork)
		case <-this.Quit.Done():
			common.MinerLoger.Debug("mining service exit")
			return
		default:
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
		this.HasNewWork = false
		this.CurrentWorkID = 0
		this.header = MinerBlockData{
			Transactions: []Transactions{},
			Parents:      []ParentItems{},
			HeaderData:   make([]byte, 0),
			TargetDiff:   &big.Int{},
			JobID:        "",
		}
		nonce := uint64(0)
		for {
			select {
			case <-this.Quit.Done():
				common.MinerLoger.Debug("mining service exit")
				return
			default:
			}
			// if has new work ,current calc stop
			if this.HasNewWork || this.ForceStop {
				break
			}
			this.Update()
			hData := make([]byte, 128)
			copy(hData[0:types.MaxBlockHeaderPayload-pow.PROOFDATA_LENGTH], this.header.HeaderBlock.BlockData())
			nonce++
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, nonce)
			copy(hData[NONCESTART:NONCEEND], b)
			h := hash.HashMeerXKeccakV1(hData[:117])
			if HashToBig(&h).Cmp(this.header.TargetDiff) <= 0 {
				headerData := BlockDataWithProof(this.header.HeaderBlock)
				copy(headerData[0:117], hData[0:117])
				common.MinerLoger.Debug(fmt.Sprintf("device #%d found hash : %s nonce:%d target:%064x", this.MinerId, h, nonce, this.header.TargetDiff))
				subm := hex.EncodeToString(headerData)
				// fmt.Println("subm", subm[:226])
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
				this.AllDiffOneShares++
				this.SubmitData <- subm
			}
		}
	}
}

func (this *MeerCrypto) GetDiff() float64 {
	s := fmt.Sprintf("%064x", this.header.TargetDiff)
	diff := float64(1)
	for i := 0; i < 64; i++ {
		if strings.ToLower(s[i:i+1]) == "f" {
			break
		}
		a, _ := strconv.ParseInt(s[i:i+1], 16, 64)
		diff *= 16 / float64(a+1)
	}
	common.MinerLoger.Debug("[current target]", "value", s, "diff", diff/1e9)
	return diff
}
func (this *MeerCrypto) Status(wg *sync.WaitGroup) {
	common.MinerLoger.Info("start listen hashrate")
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	defer wg.Done()
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Debug(fmt.Sprintf("# %d device stats service exit", this.MinerId))
			return
		case <-t.C:
			if !this.IsValid {
				return
			}
			secondsElapsed := time.Now().Unix() - this.Started
			if this.AllDiffOneShares <= 0 || secondsElapsed <= 0 {
				continue
			}
			diff := this.GetDiff()
			hashrate := float64(this.AllDiffOneShares) / float64(secondsElapsed) * diff
			// diff
			unit := "H/s"
			start := time.Unix(this.Started, 0)
			common.MinerLoger.Info(fmt.Sprintf("# %d Start time: %s  Diff: %s All Shares: %d HashRate: %s",
				this.MinerId,
				start.Format(time.RFC3339),
				common.FormatHashRate(diff, unit),
				this.AllDiffOneShares,
				common.FormatHashRate(hashrate, unit)))
		}
	}
}
