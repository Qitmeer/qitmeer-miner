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
	"unsafe"
)

type MeerCrypto struct {
	core.Device
	Work   *QitmeerWork
	header MinerBlockData
}

func (this *MeerCrypto) InitDevice() {
	this.Started = time.Now().Unix()
	common.MinerLoger.Debug(fmt.Sprintf("==============Mining MeerCrypto=============="))
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
		end := []byte{0}
		start := false
		miningEnd := false
	mining:
		for {
			// timeout TODO

			// if has new work ,current calc stop
			if this.HasNewWork || this.ForceStop { // end sig
				C.end((*C.uchar)(unsafe.Pointer(&end[0])))
				for {
					if miningEnd {
						break mining
					}
					time.Sleep(100 * time.Microsecond)
				}
			}
			this.Update()
			if !start {
				end = []byte{0}
				start = true
				go func(start, miningEnd bool) {
					defer func() {
						if r := recover(); r != nil {
							fmt.Println("Recovered in f", r)
						}
						start = false
						miningEnd = true
					}()
					hData := make([]byte, 117)
					b := make([]byte, 8)
					copy(hData[0:types.MaxBlockHeaderPayload-pow.PROOFDATA_LENGTH], this.header.HeaderBlock.BlockData())
					C.meer_pow((*C.char)(unsafe.Pointer(&hData[0])), (C.int)(len(hData)),
						(*C.char)(unsafe.Pointer(&this.header.Target2[0])),
						(*C.uchar)(unsafe.Pointer(&b[0])),
						(*C.uchar)(unsafe.Pointer(&end[0])))
					copy(hData[NONCESTART:NONCEEND], b)
					h := hash.HashMeerXKeccakV1(hData[:117])
					if HashToBig(&h).Cmp(this.header.TargetDiff) <= 0 {
						headerData := BlockDataWithProof(this.header.HeaderBlock)
						copy(headerData[0:117], hData[0:117])
						common.MinerLoger.Debug(fmt.Sprintf("device #%d found hash : %s nonce:%d target:%064x", this.MinerId, h, binary.LittleEndian.Uint64(b), this.header.TargetDiff))
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
						this.SubmitData <- subm
					}
				}(start, miningEnd)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
