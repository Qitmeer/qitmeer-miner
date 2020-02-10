package coinbase

import (
	`github.com/Qitmeer/qitmeer/core/protocol`
	`github.com/Qitmeer/qitmeer/core/types`
	`github.com/Qitmeer/qitmeer/params`
)

type ICoinbase interface {
	GetCoinbaseTx() *types.Tx
	SetRandStr(str string)
	SetExtraNonce(extraNonce uint64)
	SetPayAddr(payAddr string)
	SetHeight(h uint64)
	SetParam(p *params.Params)
	SetTotalFee(fee uint64)
	SetPackTxTotalFee(fee uint64)
	SetCoinbaseVal(val uint64)
}

type CoinbaseBase struct {
	RandStr string
	ExtraNonce uint64
	PayAddr string
	CoinbaseValue uint64
	TotalFee uint64
	PackTxTotalFee uint64
	Height uint64
	Param *params.Params
}

func (this *CoinbaseBase)SetParam(p *params.Params)  {
	this.Param = p
}

func (this *CoinbaseBase)SetHeight(h uint64)  {
	this.Height = h
}


func (this *CoinbaseBase)SetPackTxTotalFee(fee uint64)  {
	this.PackTxTotalFee = fee
}

func (this *CoinbaseBase)SetPayAddr(payAddr string)  {
	this.PayAddr = payAddr
}


func (this *CoinbaseBase)SetExtraNonce(extraNonce uint64)  {
	this.ExtraNonce = extraNonce
}

func (this *CoinbaseBase)SetRandStr(str string)  {
	this.RandStr = str
}

func (this *CoinbaseBase)SetTotalFee(fee uint64)  {
	this.TotalFee = fee
}

func (this *CoinbaseBase)SetCoinbaseVal(val uint64)  {
	this.CoinbaseValue = val
}

func GetNewCoinbaseInstance(blockVersion int,param *params.Params,payAddr string,randStr string,extraNonce uint64,height uint64,totalFee uint64,coinbaseVal uint64,packFee uint64) ICoinbase {
	var ic ICoinbase
	switch param.Net {
	case protocol.MainNet:
		switch blockVersion {
		case 0:
			ic = &Coinbase{}
		default:
			ic = &Coinbase12{}
		}
	case protocol.TestNet:
		if blockVersion <= 11{
			ic = &Coinbase{}
		}
		return &Coinbase12{}
	case protocol.MixNet:
		if blockVersion <= 17{
			ic = &Coinbase{}
		}
		return &Coinbase12{}
	case protocol.PrivNet:
		if blockVersion <= 11{
			ic = &Coinbase{}
		}
		ic = &Coinbase12{}
	default:
		return nil
	}
	ic.SetExtraNonce(extraNonce)
	ic.SetHeight(height)
	ic.SetPayAddr(payAddr)
	ic.SetParam(param)
	ic.SetRandStr(randStr)
	ic.SetCoinbaseVal(coinbaseVal)
	ic.SetTotalFee(totalFee)
	ic.SetPackTxTotalFee(packFee)
	return ic
}