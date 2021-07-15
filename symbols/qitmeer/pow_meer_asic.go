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
	"github.com/Qitmeer/qitmeer/core/types"
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
	this.Started = time.Now().Unix()
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

func (this *MeerCrypto) Mine(wg *sync.WaitGroup) {
	defer wg.Done()
	defer this.Release()
	hData := make([]byte, 117)    // header
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
		hData = make([]byte, 117) // header
		nonceBytes = make([]byte, 8)
		for {
			// if has new work ,current calc stop
			if this.HasNewWork || this.ForceStop { // end sig
				break
			}
			select {
			case <-this.Quit.Done():
				common.MinerLoger.Debug("mining service exit")
				return
			default:
				this.Update()
				copy(hData[0:types.MaxBlockHeaderPayload-pow.PROOFDATA_LENGTH], this.header.HeaderBlock.BlockData())
				if !start && fd == 0 {
					// init chips
					start = true
					copy(hData[0:types.MaxBlockHeaderPayload-pow.PROOFDATA_LENGTH], this.header.HeaderBlock.BlockData())
					fd = int(C.init_drv((C.int)(this.Cfg.OptionConfig.NumOfChips)))
				}
				nonces := make([]uint64, 0)
				// set work
				C.set_work(
					(C.int)(fd),
					(*C.uchar)(unsafe.Pointer(&hData[0])),
					(C.int)(len(hData)),
					(*C.uchar)(unsafe.Pointer(&this.header.Target2[0])),
					(C.int)(this.Cfg.OptionConfig.NumOfChips))
				chipId := make([]byte, 1)
				jobId := make([]byte, 1)
				interval := 0
				for {
					select {
					case <-this.Quit.Done():
						common.MinerLoger.Debug("mining service exit")
						return
					default:
					}
					if this.HasNewWork || this.ForceStop || interval > INTERVAL_GAP/10 { // end
						break
					}
					if fd != 0 && C.get_nonce((C.int)(fd),
						(*C.uchar)(unsafe.Pointer(&nonceBytes[0])),
						(*C.uchar)(unsafe.Pointer(&chipId[0])),
						(*C.uchar)(unsafe.Pointer(&jobId[0])),
					) {
						lastNonce := binary.LittleEndian.Uint64(nonceBytes)
						if !InUint64Array(lastNonce, nonces) {
							nonces = append(nonces, lastNonce)
							h := hash.HashMeerXKeccakV1(hData[:117])
							common.MinerLoger.Debug(fmt.Sprintf("ChipId #%d JobId #%d Found hash : %s nonce:%d target:%064x",
								chipId[0], jobId[0], h,
								lastNonce, this.header.TargetDiff))
							copy(hData[NONCESTART:NONCEEND], nonceBytes)
							if HashToBig(&h).Cmp(this.header.TargetDiff) <= 0 {
								this.AllDiffOneShares++
								headerData := BlockDataWithProof(this.header.HeaderBlock)
								copy(headerData[0:117], hData[0:117])
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
							}
						}
						time.Sleep(10 * time.Microsecond)
						interval++
					}
				}
			}
		}
	}
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
			hashrate := float64(this.AllDiffOneShares) / float64(secondsElapsed) * this.GetDiff()
			// diff
			unit := "H/s"
			common.MinerLoger.Info(fmt.Sprintf("HashRate: %s", common.FormatHashRate(hashrate, unit)))
		}
	}
}

func (this *MeerCrypto) GetDiff() float64 {
	s := hex.EncodeToString(this.header.Target2)
	diff := float64(1)
	for i := 63; i >= 0; i-- {
		if strings.ToLower(s[i:i+1]) == "f" {
			break
		}
		a, _ := strconv.ParseInt(s[i:i+1], 16, 64)
		diff *= float64(16 - a)
	}
	return diff
}

func InUint64Array(a uint64, arr []uint64) bool {
	for _, v := range arr {
		if a == v {
			return true
		}
	}
	return false
}
