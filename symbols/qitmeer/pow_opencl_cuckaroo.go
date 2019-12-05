//+build opencl,!cuda

/**
Qitmeer
james
*/
package qitmeer
import (
	`os`
	"qitmeer-miner/common"
	"qitmeer-miner/core"
	"sync"
)

type CudaCuckaroo struct {
	core.Device
	ClearBytes	[]byte
	Work                  *QitmeerWork
	header MinerBlockData
	EdgeBits            int
	Step            int
	WorkGroupSize            int
	LocalSize            int
	Nedge            int
	Edgemask            uint64
	Nonces           []uint32
}

func (this *CudaCuckaroo) InitDevice() {
	common.MinerLoger.Error("AMD Not Support CUDA!")
	os.Exit(1)
}

func (this *CudaCuckaroo) Mine(wg *sync.WaitGroup) {
	defer this.Release()
	defer wg.Done()
}

func (this *CudaCuckaroo) SubmitShare(substr chan string) {
	this.Device.SubmitShare(substr)
}