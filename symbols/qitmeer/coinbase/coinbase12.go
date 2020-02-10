package coinbase

import (
	`github.com/Qitmeer/qitmeer-miner/common`
	`github.com/Qitmeer/qitmeer/core/address`
	`github.com/Qitmeer/qitmeer/core/types`
)

type Coinbase12 struct {
	CoinbaseBase
}

func (this *Coinbase12) GetCoinbaseTx() *types.Tx {
	payToAddress, err := address.DecodeAddress(this.PayAddr)
	if err != nil {
		common.MinerLoger.Error("DecodeAddress", "error", err)
		return nil
	}
	coinbaseScript, err := standardCoinbaseScript(this.RandStr, this.Height, this.ExtraNonce)
	if err != nil {
		common.MinerLoger.Error("standardCoinbaseScript", "error", err)
		return nil
	}
	opReturnPkScript, err := standardCoinbaseOpReturn([]byte{})
	if err != nil {
		common.MinerLoger.Error("standardCoinbaseOpReturn", "error", err)
		return nil
	}
	//uit := 100000000
	subsidy := this.CoinbaseValue
	coinbaseTx, err := createCoinbaseTx(subsidy,
		coinbaseScript,
		opReturnPkScript,
		payToAddress,
		this.Param)
	if err != nil{
		common.MinerLoger.Error(err.Error())
		return nil
	}
	return coinbaseTx
}

