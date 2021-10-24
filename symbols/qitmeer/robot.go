// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qitmeer

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/core"
	"github.com/Qitmeer/qitmeer-miner/symbols/qitmeer/client"
	"github.com/Qitmeer/qitmeer-miner/symbols/qitmeer/client/cmds"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	POW_MEER_CRYPTO = "meer_crypto"
)

type PendingBlock struct {
	CoinbaseHash string
	Height       uint64
	BlockHash    string
}

type QitmeerRobot struct {
	core.MinerRobot
	Work                 QitmeerWork
	NeedGBT              chan struct{}
	Devices              []core.BaseDevice
	Stratu               *QitmeerStratum
	StratuFee            *QitmeerStratum
	AllTransactionsCount int64
	PendingBlocks        map[string]PendingBlock
	PendingLock          sync.Mutex
	SubmitLock           sync.Mutex
	WsClient             *client.Client
}

func (this *QitmeerRobot) GetPow(i int, ctx context.Context, uart_path string) core.BaseDevice {
	switch this.Cfg.NecessaryConfig.Pow {
	case POW_MEER_CRYPTO:
		deviceMiner := &MeerCrypto{}
		deviceMiner.MiningType = "meer_crypto"
		deviceMiner.UartPath = uart_path
		deviceMiner.Init(i, this.Pool, ctx, this.Cfg)
		this.Devices = append(this.Devices, deviceMiner)
		return deviceMiner

	default:
		log.Fatalln(this.Cfg.NecessaryConfig.Pow, " pow has not exist!")
	}
	return nil
}

func (this *QitmeerRobot) InitDevice(ctx context.Context) {
	this.MinerRobot.InitDevice()
	if this.Cfg.OptionConfig.CPUMiner {
		for i := 0; i < this.Cfg.OptionConfig.CpuWorkers; i++ {
			this.GetPow(i, ctx, "")
		}
	} else {
		uartPaths := strings.Split(this.Cfg.OptionConfig.UartPath, ",")
		for i := 0; i < len(uartPaths); i++ {
			if uartPaths[i] == "" {
				continue
			}
			this.GetPow(i, ctx, uartPaths[i])
		}
	}

}

// runing
func (this *QitmeerRobot) Run(ctx context.Context) {
	this.Wg = &sync.WaitGroup{}
	this.Quit = ctx
	this.InitDevice(ctx)
	connectName := "solo"
	this.Pool = false
	if this.Cfg.PoolConfig.Pool != "" { // is pool mode
		connectName = "pool"
		this.Stratu = &QitmeerStratum{}
		_ = this.Stratu.StratumConn(this.Cfg, ctx)
		this.Wg.Add(1)
		go func() {
			defer this.Wg.Done()
			this.Stratu.HandleReply()
		}()
		this.Pool = true
	}
	common.MinerLoger.Info(fmt.Sprintf("[%s miner] start", connectName))
	this.Work = QitmeerWork{}
	this.Work.Cfg = this.Cfg
	this.Work.Rpc = this.Rpc
	this.Work.stra = this.Stratu
	this.Work.Quit = this.Quit
	this.Work.WorkLock = sync.Mutex{}
	// Device Miner
	for _, dev := range this.Devices {
		dev.SetIsValid(true)
		if len(this.UseDevices) > 0 && !common.InArray(strconv.Itoa(dev.GetMinerId()), this.UseDevices) {
			dev.SetIsValid(false)
			continue
		}
		dev.SetPool(this.Pool)
		dev.InitDevice()
		this.Wg.Add(1)
		go dev.Mine(this.Wg)
		this.Wg.Add(1)
		go dev.Status(this.Wg)
	}
	//refresh work
	this.Wg.Add(1)
	go func() {
		defer this.Wg.Done()
		this.ListenWork()
	}()
	//submit work
	this.Wg.Add(1)
	go func() {
		defer this.Wg.Done()
		this.SubmitWork()
	}()
	// submit status
	this.Wg.Add(1)
	go func() {
		defer this.Wg.Done()
		this.Status()
	}()
	if this.Cfg.PoolConfig.Pool == "" {
		this.Wg.Add(1)
		go func() {
			defer this.Wg.Done()
			this.HandlePendingBlocks()
		}()
	}

	this.Wg.Wait()
}

// ListenWork
func (this *QitmeerRobot) ListenWork() {
	common.MinerLoger.Info("listen new work server")
	r := false
	first := true
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Debug("listen new work service exit")
			return
		default:
			r = false
			if this.Pool {
				r = this.Work.PoolGet() // get new work
			} else if first { // solo
				r = this.Work.Get(false) // get new work
				if r && this.Work.Block != nil {
					first = false
				}
			}
			this.NotifyWork(r)
			time.Sleep(time.Millisecond * time.Duration(this.Cfg.OptionConfig.TaskInterval))
		}
	}
}

func (this *QitmeerRobot) NotifyWork(r bool) {
	if r {
		validDeviceCount := 0
		for _, dev := range this.Devices {
			if !dev.GetIsValid() && !dev.GetIsRunning() {
				continue
			}
			dev.SetForceUpdate(false)
			validDeviceCount++
			newWork := this.Work.CopyNew()
			dev.SetNewWork(&newWork)
		}
		common.MinerLoger.Debug("New task coming", "notify device count", validDeviceCount)
	} else if this.Work.ForceUpdate {
		for _, dev := range this.Devices {
			common.MinerLoger.Debug("Task stopped by force")
			dev.SetNewWork(&this.Work)
			dev.SetForceUpdate(true)
		}
	}
}

// ListenWork
func (this *QitmeerRobot) SubmitWork() {
	common.MinerLoger.Info("listen submit block server")
	this.Wg.Add(1)
	go func() {
		defer this.Wg.Done()
		str := ""
		var logContent string
		var count int
		var arr []string
		for {
			select {
			case <-this.Quit.Done():
				close(this.SubmitStr)
				common.MinerLoger.Debug("submit service exit")
				return
			case str = <-this.SubmitStr:
				if str == "" {
					this.StaleShares++
					continue
				}
				var err error
				this.SubmitLock.Lock()
				var height, txCount, block, gbtID string
				if this.Pool {
					arr = strings.Split(str, "-")
					block = arr[0]
					err = this.Work.PoolSubmit(str)
				} else {
					//solo miner
					arr = strings.Split(str, "-")
					txCount = arr[1]
					height = arr[2]
					block = arr[0]
					gbtID = arr[3]
					err = this.Work.Submit(block, height, gbtID)
				}
				if err != nil {
					if err != ErrSameWork || err == ErrSameWork {
						if err == ErrStratumStaleWork {
							this.StaleShares++
						} else {
							this.InvalidShares++
						}
					}
					if !this.Pool {
						r := this.Work.Get(true)
						this.NotifyWork(r)
					}
					this.SubmitLock.Unlock()
				} else {
					if !this.Pool { // solo
						serializedBlock, err := hex.DecodeString(block)
						if err != nil {
							common.MinerLoger.Error(err.Error())
							continue
						}
						block, err := types.NewBlockFromBytes(serializedBlock)
						if err != nil {
							common.MinerLoger.Error(err.Error())
							continue
						}
						hei, _ := strconv.Atoi(height)
						this.PendingLock.Lock()
						this.PendingBlocks[block.Block().Transactions[0].TxHash().String()] = PendingBlock{
							Height:       uint64(hei),
							BlockHash:    block.Block().BlockHash().String(),
							CoinbaseHash: block.Block().Transactions[0].TxHash().String(),
						}
						this.PendingShares++

						txes := make([]cmds.TxConfirm, 0)
						txes = append(txes, cmds.TxConfirm{
							Txid:          block.Block().Transactions[0].TxHash().String(),
							Confirmations: int32(this.Cfg.SoloConfig.ConfirmHeight),
						})
						for _, v := range this.PendingBlocks {
							txes = append(txes, cmds.TxConfirm{
								Txid:          v.CoinbaseHash,
								Confirmations: int32(this.Cfg.SoloConfig.ConfirmHeight),
							})
						}
						common.Timeout(func() {
							if this.WsClient == nil || this.WsClient.Disconnected() {
								return
							}
							err = this.WsClient.NotifyTxsConfirmed(txes)
							if err != nil {
								common.MinerLoger.Error(err.Error())
							}
							common.MinerLoger.Info("ws block success")
						}, 1, func() {
						})
						this.PendingLock.Unlock()
						count, _ = strconv.Atoi(txCount)
						this.AllTransactionsCount += int64(count)
						logContent = fmt.Sprintf("receive block, block hash= %s, block height = %s,Including %s transactions; Received Total transactions = %d\n",
							block.Block().BlockHash().String(),
							height, txCount, this.AllTransactionsCount)
						common.MinerLoger.Info(logContent)
					} else {
						this.ValidShares++
					}
					this.SubmitLock.Unlock()
				}
			}
		}

	}()
	for _, dev := range this.Devices {
		go dev.SubmitShare(this.SubmitStr)
	}
}

// stats the submit result
func (this *QitmeerRobot) Status() {
	var valid, rejected, staleShares uint64
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Debug("global stats service exit")
			return
		case <-t.C:
			if this.Work.stra == nil && this.Work.Block == nil {
				continue
			}
			if this.Cfg.PoolConfig.Pool == "" {
				this.PendingLock.Lock()
				for i, v := range this.PendingBlocks {
					if this.Work.Block.Height > v.Height+this.Cfg.SoloConfig.NotConfirmHeight {
						common.MinerLoger.Info("[Invalid Blocks]", "block hash", v.BlockHash, "coinbase hash", v.CoinbaseHash, "height", v.Height)
						this.InvalidShares++
						this.PendingShares--
						delete(this.PendingBlocks, i)
						common.Timeout(func() {
							if this.WsClient == nil || this.WsClient.Disconnected() {
								return
							}
							txes := []cmds.TxConfirm{
								{
									Txid: v.CoinbaseHash,
								},
							}
							err := this.WsClient.RemoveTxsConfirmed(txes)
							if err != nil {
								common.MinerLoger.Error(err.Error())
							}
							common.MinerLoger.Info("ws remove success")
						}, 1, func() {
						})
					}
				}
				this.PendingLock.Unlock()
			}
			valid = this.ValidShares
			rejected = this.InvalidShares
			staleShares = this.StaleShares
			if this.Pool {
				valid = this.Stratu.ValidShares
				rejected = this.Stratu.InvalidShares
				staleShares = this.Stratu.StaleShares
			}
			this.Cfg.OptionConfig.Accept = int(valid)
			this.Cfg.OptionConfig.Reject = int(rejected)
			this.Cfg.OptionConfig.Stale = int(staleShares)
			total := valid + rejected + staleShares + this.PendingShares
			common.MinerLoger.Info(fmt.Sprintf("Global stats: Accepted: %v,Pending: %v,Stale: %v, Rejected: %v, Total: %v",
				valid,
				this.PendingShares,
				staleShares,
				rejected,
				total,
			))
		}
	}
}

// stats the submit result
func (this *QitmeerRobot) HandlePendingBlocks() {
	this.WsConnect()
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Debug("Exit Websocket")
			if this.WsClient != nil && !this.WsClient.Disconnected() {
				this.WsClient.Shutdown()
			}
			return
		default:
			if this.WsClient == nil || this.WsClient.Disconnected() {
				this.WsConnect()
			}
		}
	}
}

func (this *QitmeerRobot) WsConnect() {
	var err error
	ntfnHandlers := client.NotificationHandlers{
		OnTxConfirm: func(txConfirm *cmds.TxConfirmResult) {
			this.PendingLock.Lock()
			common.MinerLoger.Info("OnTxConfirm", "tx", txConfirm.Tx, "confirms", txConfirm.Confirms, "order", txConfirm.Order)
			if _, ok := this.PendingBlocks[txConfirm.Tx]; ok && txConfirm.Confirms >= this.Cfg.SoloConfig.ConfirmHeight {
				//
				if _, ok := this.PendingBlocks[txConfirm.Tx]; ok {
					this.PendingShares--
					this.ValidShares++
					delete(this.PendingBlocks, txConfirm.Tx)
				}
			} else {
				if _, ok := this.PendingBlocks[txConfirm.Tx]; ok {
					this.PendingShares--
					this.InvalidShares++
					delete(this.PendingBlocks, txConfirm.Tx)
				}
			}
			this.PendingLock.Unlock()
		},
		OnBlockConnected: func(hash *hash.Hash, height, order int64, t time.Time, txs []*types.Transaction) {
			go func() {
				this.SubmitLock.Lock()
				r := this.Work.Get(false)
				if this.Work.Block != nil {
					common.MinerLoger.Info("New Block Coming", "height", height, "gbt height", this.Work.Block.Height, "cur height", common.CurrentHeight)
				}
				this.NotifyWork(r)
				this.SubmitLock.Unlock()
			}()
		},
		OnNodeExit: func(p *cmds.NodeExitNtfn) {
			common.MinerLoger.Debug("OnNodeExit")
		},
	}
	protocol := "ws"
	if !this.Cfg.SoloConfig.NoTLS {
		protocol = "wss"
	}
	url := this.Cfg.SoloConfig.RPCServer
	noTls := this.Cfg.SoloConfig.NoTLS
	if strings.Contains(this.Cfg.SoloConfig.RPCServer, "://") {
		arr := strings.Split(url, "://")
		url = arr[1]
		protocol = "ws"
		if arr[0] == "https" {
			noTls = false
			arr[0] = "wss"
		}
	}
	connCfg := &client.ConnConfig{
		Host:       url,
		Endpoint:   protocol,
		User:       this.Cfg.SoloConfig.RPCUser,
		Pass:       this.Cfg.SoloConfig.RPCPassword,
		DisableTLS: noTls,
	}
	if !connCfg.DisableTLS {
		certs, err := ioutil.ReadFile(this.Cfg.SoloConfig.RPCCert)
		if err != nil {
			common.MinerLoger.Error("rpccert= need config", err.Error())
			time.Sleep(10 * time.Second)
			return
		}
		connCfg.Certificates = certs
	}

	this.WsClient, err = client.New(connCfg, &ntfnHandlers)
	if err != nil {
		common.MinerLoger.Error(err.Error())
		return
	}
	// Register for block connect and disconnect notifications.
	if err := this.WsClient.NotifyBlocks(); err != nil {
		common.MinerLoger.Error(err.Error())
		return
	}
}
