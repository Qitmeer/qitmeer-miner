// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qitmeer

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/core"
	"github.com/Qitmeer/qitmeer-miner/stats_server"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	POW_MEER_CRYPTO = "meer_crypto"
)

type QitmeerRobot struct {
	core.MinerRobot
	Work                 QitmeerWork
	Devices              []core.BaseDevice
	Stratu               *QitmeerStratum
	StratuFee            *QitmeerStratum
	AllTransactionsCount int64
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
			if uartPaths[0] == "" {
				continue
			}
			this.GetPow(i, ctx, uartPaths[0])
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

	// http server stats
	if this.Cfg.OptionConfig.StatsServer != "" {
		this.Wg.Add(1)
		go func() {
			defer this.Wg.Done()
			stats_server.HandleRouter(this.Cfg, this.Devices)
		}()
	}
	this.Wg.Wait()
}

// ListenWork
func (this *QitmeerRobot) ListenWork() {
	common.MinerLoger.Info("listen new work server")
	t := time.NewTicker(time.Second * time.Duration(this.Cfg.OptionConfig.TaskInterval))
	isFirst := true
	defer t.Stop()
	r := false
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Debug("listen new work service exit")
			return
		case <-t.C:
			r = false
			if this.Pool {
				r = this.Work.PoolGet() // get new work
			} else {
				r = this.Work.Get() // get new work
			}
			if r {
				common.MinerLoger.Debug("New task coming")
				validDeviceCount := 0
				for _, dev := range this.Devices {
					if !dev.GetIsValid() {
						continue
					}
					dev.SetForceUpdate(false)
					validDeviceCount++
					newWork := this.Work.CopyNew()
					dev.SetNewWork(&newWork)
				}
				if validDeviceCount <= 0 {
					common.MinerLoger.Error("There is no valid device to mining,please check your config!")
					os.Exit(1)
				}
				if isFirst {
					isFirst = false
				}
			} else if this.Work.ForceUpdate {
				for _, dev := range this.Devices {
					common.MinerLoger.Debug("Task stopped by force")
					dev.SetNewWork(&this.Work)
					dev.SetForceUpdate(true)
				}
			}
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
				var height, txCount, block string
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
					err = this.Work.Submit(block)
				}
				if err != nil {
					if err != ErrSameWork || err == ErrSameWork {
						if err == ErrStratumStaleWork {
							this.StaleShares++
						} else {
							this.InvalidShares++
						}
					}
				} else {
					this.ValidShares++
					if !this.Pool {
						count, _ = strconv.Atoi(txCount)
						this.AllTransactionsCount += int64(count)
						logContent = fmt.Sprintf("receive block, block height = %s,Including %s transactions; Received Total transactions = %d\n",
							height, txCount, this.AllTransactionsCount)
						common.MinerLoger.Info(logContent)
					}
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
	for {
		select {
		case <-this.Quit.Done():
			common.MinerLoger.Debug("global stats service exit")
			return
		default:
			if this.Work.stra == nil && this.Work.Block == nil {
				common.Usleep(20)
				continue
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
			total := valid + rejected + staleShares
			common.MinerLoger.Info(fmt.Sprintf("Global stats: Accepted: %v,Stale: %v, Rejected: %v, Total: %v",
				valid,
				staleShares,
				rejected,
				total,
			))
			common.Usleep(20)
		}
	}
}
