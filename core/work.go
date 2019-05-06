/**
	HLC FOUNDATION
	james
 */
package core

import (
	"hlc-miner/common"
	"sync"
	"os"
)

//standard work template
type Work struct {
	Cfg *common.Config
	Rpc *common.RpcClient
	Clean bool
	sync.Mutex
	Quit chan os.Signal
	Started uint32
	LastSub string //last submit string
}

//GetBlockTemplate
func (this *Work) Get ()  {

}

//Submit
func (this *Work) Submit ()  {

}

// pool get work
func (this *Work) PoolGet ()  {

}

//pool submit work
func (this *Work) PoolSubmit ()  {

}