// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qitmeer

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/go-opencl/cl"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	"log"
	"qitmeer-miner/common"
	"qitmeer-miner/core"
	"qitmeer-miner/stats_server"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	POW_DOUBLE_BLAKE2B = "blake2bd"
	POW_CUCKROO = "cuckaroo"
	POW_CUCKTOO = "cuckatoo"
)
type QitmeerRobot struct {
	core.MinerRobot
	Work                 QitmeerWork
	Devices              []core.BaseDevice
	Stratu               *QitmeerStratum
	AllTransactionsCount int64
}

func (this *QitmeerRobot)GetPow(i int ,device *cl.Device) core.BaseDevice{
	switch this.Cfg.NecessaryConfig.Pow {
	case POW_CUCKROO:
		deviceMiner := &Cuckaroo{}
		deviceMiner.Init(i,device,this.Pool,this.Quit,this.Cfg)
		this.Devices = append(this.Devices,deviceMiner)
		return deviceMiner
	case POW_CUCKTOO:
	case POW_DOUBLE_BLAKE2B:
		deviceMiner := &Blake2bD{}
		deviceMiner.Init(i,device,this.Pool,this.Quit,this.Cfg)
		this.Devices = append(this.Devices,deviceMiner)
		return deviceMiner

	default:
		log.Fatalln(this.Cfg.NecessaryConfig.Pow," pow has not exist!")
	}
	return nil
}

func (this *QitmeerRobot)InitDevice()  {
	this.MinerRobot.InitDevice()
	for i, device := range this.ClDevices {
		deviceMiner := this.GetPow(i ,device)
		if deviceMiner == nil{
			return
		}
	}
}

// runing
func (this *QitmeerRobot)Run() {
	this.InitDevice()
	//mining service
	this.Wg.Add(1)
	go func() {
		defer this.Wg.Done()
		for{
			if this.Cfg.OptionConfig.Restart == 1{
				for _,dev := range this.Devices{
					if dev.GetStart() > 0 {
						dev.Release()
					}
				}
				common.MinerLoger.Debug("miner will start after 5s due to close the thread before!")
				time.Sleep(5*time.Second)
			}
			this.Cfg.OptionConfig.Restart = 0
			wg := &sync.WaitGroup{}
			connectName := "solo"
			this.Pool = false
			if this.Cfg.PoolConfig.Pool != ""{ //is pool mode
				connectName = "pool"
				this.Stratu = &QitmeerStratum{}
				err := this.Stratu.StratumConn(this.Cfg)
				if err != nil {
					common.MinerLoger.Error(err.Error())
					time.Sleep(1*time.Second)
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					this.Stratu.HandleReply()
				}()
				this.Pool = true
			}
			common.MinerLoger.Infof("%s miner start",connectName)
			this.Work = QitmeerWork{}
			this.Work.Cfg = this.Cfg
			this.Work.Rpc = this.Rpc
			this.Work.stra = this.Stratu
			// Device Miner
			for _,dev := range this.Devices{
				dev.SetIsValid(true)
				dev.InitDevice()
				dev.SetPool(this.Pool)
				if this.Cfg.OptionConfig.UseDevices != ""{
					this.UseDevices = strings.Split(this.Cfg.OptionConfig.UseDevices,",")
				}
				if len(this.UseDevices) > 0 && !common.InArray(strconv.Itoa(dev.GetMinerId()),this.UseDevices){
					dev.SetIsValid(false)
				}
				go dev.Mine()
				go dev.Status()
			}
			//refresh work
			wg.Add(1)
			go func(){
				defer wg.Done()
				this.ListenWork()
			}()
			//submit work
			wg.Add(1)
			go func(){
				defer wg.Done()
				this.SubmitWork()
			}()
			//submit status
			wg.Add(1)
			go func(){
				defer wg.Done()
				this.Status()
			}()
			wg.Wait()
		}
	}()

	//http server stats
	this.Wg.Add(1)
	go func(){
		defer this.Wg.Done()
		stats_server.HandleRouter(this.Cfg,this.Devices)
	}()
	this.Wg.Wait()
}

// ListenWork
func (this *QitmeerRobot)ListenWork() {
	common.MinerLoger.Info("listen new work server")
	time.Sleep(1*time.Second)
	for {
		if this.Cfg.OptionConfig.Restart == 1{
			common.MinerLoger.Info("listen server restart")
			return
		}
		select {
		case <-this.Quit:
			return
		default:
			var r = false
			if this.Pool {
				r = this.Work.PoolGet() // get new work
			} else {
				r = this.Work.Get() // get new work
			}
			if r {
				for _, dev := range this.Devices {
					dev.SetNewWork(&this.Work)
				}
			}
		}
		time.Sleep(1*time.Second)
	}
}

// ListenWork
func (this *QitmeerRobot)SubmitWork() {
	common.MinerLoger.Info("listen submit block server")
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		str := ""
		for{
			if this.Cfg.OptionConfig.Restart == 1{
				common.MinerLoger.Info("submit server restart")
				return
			}
			select {
			case <-this.Quit:
				return
			case str = <-this.SubmitStr:
				if str == ""{
					atomic.AddUint64(&this.StaleShares, 1)
					continue
				}
				var err error
				var height ,txCount ,block string
				if this.Pool {
					arr := strings.Split(str,"-")
					block = arr[0]
					err = this.Work.PoolSubmit(str)
				} else {
					//solo miner
					arr := strings.Split(str,"-")
					txCount = arr[1]
					height = arr[2]
					block = arr[0]
					err = this.Work.Submit(block)
				}
				if err != nil{
					if err == ErrStratumStaleWork || err == ErrSameWork{
						atomic.AddUint64(&this.StaleShares, 1)
					} else{
						atomic.AddUint64(&this.InvalidShares, 1)
					}
				} else {
					byt ,_:= hex.DecodeString(block)
					common.MinerLoger.Infof("[Found hash and submit]%s",hash.DoubleHashH(byt[0:Blake2bBlockLength]))
					atomic.AddUint64(&this.ValidShares, 1)
					if !this.Pool{
						count ,_ := strconv.Atoi(txCount)
						this.AllTransactionsCount += int64(count)
						logContent := fmt.Sprintf("%s,receive block, block height = %s,Including %s transactions(not contain coinbase tx); Received Total transactions = %d\n",
							time.Now().Format("2006-01-02 03:04:05 PM"),height,txCount,this.AllTransactionsCount)
						common.MinerLoger.Info(logContent)
					}
				}
			default:

			}
		}

	}()
	for _,dev := range this.Devices{
		go dev.SubmitShare(this.SubmitStr)
	}
	wg.Wait()
}

// stats the submit result
func (this *QitmeerRobot)Status()  {
	t := time.NewTicker(time.Second * 30)
	defer t.Stop()
	for {
		if this.Cfg.OptionConfig.Restart == 1{
			common.MinerLoger.Debug("stats restart")
			return
		}
		select {
		case <-this.Quit:
			return
		case <-t.C:
			if this.Work.stra == nil && this.Work.Block.Height == 0{
				continue
			}
			valid := atomic.LoadUint64(&this.ValidShares)
			rejected := atomic.LoadUint64(&this.InvalidShares)
			staleShares := atomic.LoadUint64(&this.StaleShares)
			if this.Pool{
				valid = atomic.LoadUint64(&this.Stratu.ValidShares)
				rejected = atomic.LoadUint64(&this.Stratu.InvalidShares)
				staleShares = atomic.LoadUint64(&this.Stratu.StaleShares)
			}
			this.Cfg.OptionConfig.Accept = int(valid)
			this.Cfg.OptionConfig.Reject = int(rejected)
			this.Cfg.OptionConfig.Stale = int(staleShares)
			total := valid + rejected + staleShares
			common.MinerLoger.Infof("Global stats: Accepted: %v,Stale: %v, Rejected: %v, Total: %v",
				valid,
				staleShares,
				rejected,
				total,
			)
		default:

		}
	}
}