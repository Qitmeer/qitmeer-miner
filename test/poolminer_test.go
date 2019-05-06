package test

import (
	"testing"
	"log"
	"runtime"
	"os"
	"os/signal"
)

func TestPool(t *testing.T){
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
	robotminer = GetRobot(cfg,"pool")
	if robotminer == nil{
		log.Fatalln("please check the coin in the README.md! if this coin is supported")
		return
	}
	robotminer.Run()
}
