// Copyright (c) 2019 The halalchain developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"runtime"
	"hlc-miner/core"
	"hlc-miner/common"
	"log"
	"hlc-miner/symbols/hlc"
	"os"
	"os/signal"
	"time"
	"strings"
)
var robotminer core.Robot

//init the config file
func init(){
	cfg, _, err := common.LoadConfig()
	if err != nil {
		log.Fatal("Config file error,please check it.【",err,"】")
		return
	}
	//test config
	//init miner robot
	robotminer = GetRobot(cfg)
}

func main()  {
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Got Control+C, exiting...")
		os.Exit(0)
	}()
	if robotminer == nil{
		log.Fatalln("please check the coin in the README.md! if this coin is supported")
		return
	}
	robotminer.Run()
}

//get current coin miner
func GetRobot(cfg *common.Config) core.Robot {
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
