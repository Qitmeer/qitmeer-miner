//+build opencl,!cuda

/**
Qitmeer
james
*/
package qitmeer

import (
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/core"
	"os"
	"sync"
)

type CudaCuckaroom struct {
	core.Device
	ClearBytes    []byte
	Work          *QitmeerWork
	header        MinerBlockData
	EdgeBits      int
	Step          int
	WorkGroupSize int
	LocalSize     int
	Nedge         int
	Edgemask      uint64
	Nonces        []uint32
}

func (this *CudaCuckaroom) InitDevice() {
	common.MinerLoger.Error("AMD Not Support CUDA!")
	os.Exit(1)
}

func (this *CudaCuckaroom) Mine(wg *sync.WaitGroup) {
	defer this.Release()
	defer wg.Done()
}

func (this *CudaCuckaroom) SubmitShare(substr chan string) {
	this.Device.SubmitShare(substr)
}
