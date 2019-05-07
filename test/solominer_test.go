package test

import (
	"testing"
	"log"
	"hlc-miner/common"
	"hlc-miner/core"
	"strings"
	"hlc-miner/symbols/hlc"
	"time"
	"runtime"
	"os"
	"os/signal"
)
var robotminer core.Robot
var cfg *common.Config
//init the config file
func init(){
	var err error
	cfg, _, err = common.LoadConfig()
	if err != nil {
		log.Fatal("Config file error,please check it.【",err,"】")
		return
	}
	//
}

func TestSolo(t *testing.T){
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Got Control+C, exiting...")
		os.Exit(0)
	}()
	//init miner robot
	robotminer = GetRobot(cfg,"solo")
	if robotminer == nil{
		log.Fatalln("please check the coin in the README.md! if this coin is supported")
		return
	}
	robotminer.Run()
}

//get current coin miner
func GetRobot(cfg *common.Config,mode string) core.Robot {
	switch strings.ToUpper(cfg.Symbol) {
	case core.SYMBOL_NOX:
		r := &hlc.HLCRobot{}
		r.Cfg = cfg
		r.Started = uint32(time.Now().Unix())
		r.Rpc = &common.RpcClient{Cfg:cfg,}
		r.SubmitStr = make(chan string,0)
		return r
	}
	return nil
}
