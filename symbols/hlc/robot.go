/**
	HLC FOUNDATION
	james
 */
package hlc

import (
	"hlc-miner/core"
	"log"
	"sync"
	"time"
	"math"
	"fmt"
	"sync/atomic"
	"strings"
	"strconv"
	"hlc-miner/common"
)

type HLCRobot struct {
	core.MinerRobot
	Work HLCWork
	Devices 	  []*HLCDevice
	Stratu      *HLCStratum
	AllTransactionsCount     int64
}

func (this *HLCRobot)InitDevice()  {
	this.MinerRobot.InitDevice()
	this.Devices = make([]*HLCDevice,0)
	for i, device := range this.ClDevices {
		deviceMiner := &HLCDevice{
			NewWork:make(chan HLCWork),
		}
		deviceMiner.MinerId = uint32(i)
		deviceMiner.Cfg=this.Cfg
		deviceMiner.DeviceName=device.Name()
		deviceMiner.ClDevice=device
		deviceMiner.CurrentWorkID=0
		deviceMiner.IsValid=true
		deviceMiner.Pool=this.Pool
		deviceMiner.SubmitData=make(chan string,0)
		deviceMiner.Started=uint32(time.Now().Unix())
		deviceMiner.GlobalItemSize= int(math.Exp2(float64(this.Cfg.Intensity)))
		deviceMiner.Quit=this.Quit
		log.Println(fmt.Sprintf("Found Can Mining Device %d : ",i),deviceMiner.DeviceName)
		deviceMiner.AllDiffOneShares = 0
		this.Devices = append(this.Devices,deviceMiner)
	}
}

// runing
func (this *HLCRobot)Run() {
	log.Println("miner start")
	if this.Cfg.Pool != ""{ //is pool mode
		this.Stratu = &HLCStratum{}
		err := this.Stratu.StratumConn(this.Cfg)
		if err != nil {
			log.Fatalln(err)
			return
		}
		go this.Stratu.HandleReply()
		this.Pool = true
	}
	this.Work = HLCWork{}
	this.Work.Cfg = this.Cfg
	this.Work.Rpc = this.Rpc
	this.Work.stra = this.Stratu
	this.InitDevice()
	// Device Miner
	for _,dev := range this.Devices{
		dev.InitDevice()
		go dev.Mine()
		go dev.Status()
	}
	//refresh work
	this.Wg.Add(1)
	go func(){
		defer this.Wg.Done()
		this.ListenWork()
	}()
	//submit work
	this.Wg.Add(1)
	go func(){
		defer this.Wg.Done()
		this.SubmitWork()
	}()
	//submit status
	this.Wg.Add(1)
	go func(){
		defer this.Wg.Done()
		this.Status()
	}()
	this.Wg.Wait()
}

// ListenWork
func (this *HLCRobot)ListenWork() {
	log.Println("listen new work server")
	time.Sleep(1*time.Second)
	for {
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
				//log.Println("has work")
				//this.Work.StartWork = true
				for _, d := range this.Devices {
					if !d.IsValid{
						continue
					}
					d.HasNewWork = true
					d.NewWork <- this.Work
				}
			} else{
				//log.Println("not has work")
			}
		}
		time.Sleep(1*time.Second)
	}
}

// ListenWork
func (this *HLCRobot)SubmitWork() {
	log.Println("listen submit block server")
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		str := ""
		for{
			select {
			case <-this.Quit:
				return
			case str = <-this.SubmitStr:
				if str == ""{
					atomic.AddUint64(&this.StaleShares, 1)
					continue
				}
				var err error
				var height ,txCount string
				if this.Pool {
					err = this.Work.PoolSubmit(str)
				} else {
					//solo miner
					arr := strings.Split(str,"-")
					txCount = arr[1]

					height = arr[2]
					err = this.Work.Submit(arr[0])
				}
				if err != nil{
					if err != ErrSameWork{
						//log.Println("【submit error】:",err)
						if err == ErrStratumStaleWork{
							atomic.AddUint64(&this.StaleShares, 1)
						} else{
							log.Println(err)
							atomic.AddUint64(&this.InvalidShares, 1)
						}
					}
				} else {
					atomic.AddUint64(&this.ValidShares, 1)
					count ,_ := strconv.Atoi(txCount)
					this.AllTransactionsCount += int64(count)
					logContent := fmt.Sprintf("%s,receive block, block height = %s,Including %s transactions; Received Total transactions = %d\n",
						time.Now().Format("2006-01-02 03:04:05 PM"),height,txCount,this.AllTransactionsCount)
					common.AppendToFile("/tmp/miner.log",logContent)
				}
			}
		}

	}()
	for _,dev := range this.Devices{
		go dev.SubmitShare(this.SubmitStr)
	}
	wg.Wait()
}

// stats the submit result
func (this *HLCRobot)Status()  {
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	for {
		select {
		case <-this.Quit:
			return
		case <-t.C:
			valid := atomic.LoadUint64(&this.ValidShares)
			rejected := atomic.LoadUint64(&this.InvalidShares)
			staleShares := atomic.LoadUint64(&this.StaleShares)
			if this.Pool{
				valid = atomic.LoadUint64(&this.Stratu.ValidShares)
				rejected = atomic.LoadUint64(&this.Stratu.InvalidShares)
				staleShares = atomic.LoadUint64(&this.Stratu.StaleShares)
			}
			total := valid + rejected + staleShares
			log.Printf("Global stats: Accepted: %v,Stale: %v, Rejected: %v, Total: %v",
				valid,
				staleShares,
				rejected,
				total,
			)
		}
	}
}