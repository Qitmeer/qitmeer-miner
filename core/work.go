// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package core

import (
	"context"
	"github.com/Qitmeer/qitmeer-miner/common"
	"sync"
)

type BaseWork interface {
	Get(force bool) bool
	Submit(subm, height, gbtID string) error
	PoolGet() bool
	PoolSubmit(subm string) error
}

//standard work template
type Work struct {
	Cfg   *common.GlobalConfig
	Rpc   *common.RpcClient
	Clean bool
	sync.Mutex
	Quit        context.Context
	Started     uint32
	GetWorkTime int64
	LastSub     string //last submit string
}
