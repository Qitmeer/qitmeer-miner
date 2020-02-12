package coinbase

import (
	`github.com/Qitmeer/qitmeer/core/types`
)

type Coinbase086 struct {
	CoinbaseBase
}

//in qitmeer 0.8.6 the getblocktemplate result about coinbasevalue do not contain tx fee (only contain subsidy)
//the miner coinbase tx need not contain tx fee (only just subsidy)
func (this *Coinbase086) GetCoinbaseTx() *types.Tx {
	return this.CalcCoinbaseTx(this.CoinbaseValue)
}

