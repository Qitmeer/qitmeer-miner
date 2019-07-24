// Copyright (c) 2019 The halalchain developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package core

import (
	"qitmeer-miner/common"
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
	Cfg *common.GlobalConfig
	Rpc *common.RpcClient
	Clean bool
	sync.Mutex
	Quit chan os.Signal
	Started uint32
	LastSub string //last submit string
}