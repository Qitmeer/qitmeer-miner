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

type BaseWork interface {
	Get() bool
	Submit(subm string) error
	PoolGet() bool
	PoolSubmit(subm string) error
}

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