//+build asic,!cpu

/**
Qitmeer
james
*/
package qitmeer

/*
#include "../../asic/meer/main.h"
#include "../../asic/meer/main.c"
#include "../../asic/meer/algo_meer.c"
#include "../../asic/meer/meer.h"
#include "../../asic/meer/meer_drv.c"
#include "../../asic/meer/meer_drv.h"
#include "../../asic/meer/uart.h"
#include "../../asic/meer/uart.c"
#cgo CFLAGS: -Wno-unused-result
#cgo CFLAGS: -Wno-int-conversion
*/
import "C"
import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/core"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type MeerCrypto struct {
	core.Device
	Work   *QitmeerWork
	header MinerBlockData
}

func (this *MeerCrypto) InitDevice() {
	common.MinerLoger.Debug("==============Mining MeerCrypto ==============", "chips num", this.Cfg.OptionConfig.NumOfChips)
}

const INTERVAL_GAP = 2049630

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

type MiningResultItem struct {
	Nonce  uint64
	JobId  byte
	ChipId byte
}

type Work struct {
	ChipId    byte
	Header    []byte
	Target    *big.Int
	SubmitStr string
}

type MiningResult map[uint64]MiningResultItem

func (this *MeerCrypto) Mine(wg *sync.WaitGroup) {
	defer func() {
		// recover from panic caused by writing to a closed channel
		if r := recover(); r != nil {
			common.MinerLoger.Debug(fmt.Sprintf("# %d miner service exit", this.MinerId))
			return
		}
		common.MinerLoger.Debug(fmt.Sprintf("# %d miner service exit", this.MinerId))
	}()
	defer wg.Done()
	defer this.Release()
	nonceBytes := make([]byte, 8) // nonce bytes
	start := false
	fd := 0
	var w core.BaseWork
	defer func() {
		if fd > 0 {
			C.meer_drv_deinit((C.int)(fd))
		}
	}()
	this.Started = time.Now().Unix()
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
		for !this.HasNewWork && !this.ForceStop {
			// if has new work ,current calc stop
			select {
			case <-this.Quit.Done():
				common.MinerLoger.Debug("mining service exit")
				return
			default:
				if !start && fd == 0 {
					// init chips
					start = true
					fd = int(C.init_drv((C.int)(this.Cfg.OptionConfig.NumOfChips)))
				}
				nonces := MiningResult{}
				works := map[byte]Work{}
				for j := 1; j <= this.Cfg.OptionConfig.NumOfChips; j++ {
					this.Update()
					works[byte(j)] = Work{
						ChipId:    byte(j),
						Header:    make([]byte, 117),
						Target:    this.header.TargetDiff,
						SubmitStr: this.GetSubmitStr(),
					}
					copy(works[byte(j)].Header[0:117], this.header.HeaderBlock.BlockData())
					C.set_work(
						(C.int)(fd),
						(*C.uchar)(unsafe.Pointer(&works[byte(j)].Header[0])),
						(C.int)(len(works[byte(j)].Header)),
						(*C.uchar)(unsafe.Pointer(&this.header.Target2[0])),
						(C.int)(j))
				}
				// set work
				start := time.Now().Unix()
				// 10 mill second next task
				for time.Now().Unix()-start < int64(this.Cfg.OptionConfig.Timeout) && !this.HasNewWork && !this.ForceStop {
					select {
					case <-this.Quit.Done():
						common.MinerLoger.Debug("mining service exit")
						return
					default:
					}

					chipId := make([]byte, 1)
					jobId := make([]byte, 1)
					nonceBytes = make([]byte, 8)
					if fd != 0 && C.get_nonce((C.int)(fd),
						(*C.uchar)(unsafe.Pointer(&nonceBytes[0])),
						(*C.uchar)(unsafe.Pointer(&chipId[0])),
						(*C.uchar)(unsafe.Pointer(&jobId[0])),
					) {
						if chipId[0] < 1 || chipId[0] > byte(this.Cfg.OptionConfig.NumOfChips) {
							time.Sleep(10 * time.Millisecond)
							continue
						}
						cwork := works[chipId[0]]
						lastNonce := binary.LittleEndian.Uint64(nonceBytes)
						if _, ok := nonces[lastNonce]; !ok {
							nonces[lastNonce] = MiningResultItem{
								Nonce:  lastNonce,
								JobId:  jobId[0],
								ChipId: chipId[0],
							}
							copy(cwork.Header[NONCESTART:NONCEEND], nonceBytes)
							h := hash.HashMeerXKeccakV1(cwork.Header[:117])
							common.MinerLoger.Debug(fmt.Sprintf("ChipId #%d JobId #%d Found hash : %s nonce:%s target:%064x",
								chipId[0], jobId[0], h,
								hex.EncodeToString(nonceBytes), cwork.Target))
							if HashToBig(&h).Cmp(cwork.Target) <= 0 {
								this.AllDiffOneShares++
								this.SubmitData <- cwork.ReplaceNonce(nonceBytes)
							}
						} else {
							common.MinerLoger.Debug(fmt.Sprintf("[DUP Shares]ChipId #%d JobId #%d nonce:%d  Last ChipId: %d Last JobId :%d ",
								chipId[0], jobId[0],
								lastNonce,
								nonces[lastNonce].ChipId, nonces[lastNonce].JobId))
						}
						time.Sleep(10 * time.Millisecond)
					}
				}
			}
		}
	}
}

func (this *MeerCrypto) GetSubmitStr() string {
	headerData := BlockDataWithProof(this.header.HeaderBlock)
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
	return subm
}

func (this *MeerCrypto) Status(wg *sync.WaitGroup) {
	common.MinerLoger.Info("start listen hashrate")
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	defer wg.Done()
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Debug("device stats service exit")
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
			common.MinerLoger.Info(fmt.Sprintf("Start time: %s  Diff: %s All Shares: %d HashRate: %s",
				start.Format(time.RFC3339),
				common.FormatHashRate(diff, unit),
				this.AllDiffOneShares,
				common.FormatHashRate(hashrate, unit)))
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

func (this *Work) ReplaceNonce(nonce []byte) string {
	arr := strings.Split(this.SubmitStr, "-")
	b, err := hex.DecodeString(arr[0])
	if err != nil {
		return this.SubmitStr
	}
	copy(b[0:117], this.Header)
	copy(b[NONCESTART:NONCEEND], nonce)
	arr[0] = hex.EncodeToString(b)
	return strings.Join(arr, "-")
}
