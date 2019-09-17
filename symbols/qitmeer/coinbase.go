// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qitmeer

import (
	"encoding/hex"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	"github.com/Qitmeer/qitmeer-lib/core/address"
	"github.com/Qitmeer/qitmeer-lib/core/types"
	"github.com/Qitmeer/qitmeer-lib/engine/txscript"
	"github.com/Qitmeer/qitmeer-lib/params"
	"github.com/google/uuid"
	"qitmeer-miner/common"
	"sort"
)


// standardCoinbaseOpReturn creates a standard OP_RETURN output to insert into
// coinbase to use as extranonces. The OP_RETURN pushes 32 bytes.
func standardCoinbaseOpReturn(enData []byte) ([]byte, error) {
	if len(enData) == 0 {
		return nil,nil
	}
	extraNonceScript, err := txscript.GenerateProvablyPruneableOut(enData)
	if err != nil {
		return nil, err
	}
	return extraNonceScript, nil
}

func standardCoinbaseScript(randStr string,nextBlockHeight uint64, extraNonce uint64) ([]byte, error) {
	uniqueStr := uuid.New()
	return txscript.NewScriptBuilder().AddInt64(int64(nextBlockHeight)).
		AddInt64(int64(extraNonce)).AddData([]byte(randStr)).AddData([]byte(uniqueStr.String())).
		Script()
}

// CalcBlockTaxSubsidy calculates the subsidy for the organization address in the
// coinbase.
func CalcBlockTaxSubsidy(coinbaseVal uint64, params *params.Params) uint64 {
	_,_,tax:=calcBlockProportion(coinbaseVal,params)
	return tax
}

func calcSubsidyByCoinBase(coinbaseVal uint64, params *params.Params) uint64{
	workPro := float64(params.WorkRewardProportion)
	proportions := float64(params.TotalSubsidyProportions())
	subsidy := float64(coinbaseVal) * proportions / workPro
	return uint64(subsidy)
}

func calcBlockProportion(coinbaseVal uint64, params *params.Params) (uint64,uint64,uint64) {
	subsidy := calcSubsidyByCoinBase(coinbaseVal,params)
	workPro := float64(params.WorkRewardProportion)
	stakePro:= float64(params.StakeRewardProportion)
	proportions := float64(params.TotalSubsidyProportions())
	work:=uint64(workPro/proportions*float64(subsidy))
	stake:=uint64(stakePro/proportions*float64(subsidy))
	tax:=subsidy-work-stake
	return work,stake,tax
}

// createCoinbaseTx returns a coinbase transaction paying an appropriate subsidy
// based on the passed block height to the provided address.  When the address
// is nil, the coinbase transaction will instead be redeemable by anyone.
//
// See the comment for NewBlockTemplate for more information about why the nil
// address handling is useful.
func createCoinbaseTx(coinBaseVal uint64,coinbaseScript []byte, opReturnPkScript []byte, addr types.Address, params *params.Params) (*types.Tx, error) {
	tx := types.NewTransaction()
	tx.AddTxIn(&types.TxInput{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOut: *types.NewOutPoint(&hash.Hash{},
			types.MaxPrevOutIndex ),
		Sequence:        types.MaxTxInSequenceNum,
		SignScript:      coinbaseScript,
	})

	hasTax:=false
	if params.BlockTaxProportion > 0 &&
		len(params.OrganizationPkScript) > 0{
		hasTax=true
	}
	// Create a coinbase with correct block subsidy and extranonce.
	subsidy := coinBaseVal
	tax := CalcBlockTaxSubsidy(coinBaseVal, params)
	// output
	// Create the script to pay to the provided payment address if one was
	// specified.  Otherwise create a script that allows the coinbase to be
	// redeemable by anyone.
	var pksSubsidy []byte
	var err error
	if addr != nil {
		pksSubsidy, err = txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
	} else {
		scriptBuilder := txscript.NewScriptBuilder()
		pksSubsidy, err = scriptBuilder.AddOp(txscript.OP_TRUE).Script()
		if err != nil {
			return nil, err
		}
	}
	if !hasTax {
		subsidy+=uint64(tax)
		tax=0
	}
	// Subsidy paid to miner.
	tx.AddTxOut(&types.TxOutput{
		Amount:   subsidy,
		PkScript: pksSubsidy,
	})

	// Tax output.
	if hasTax {
		tx.AddTxOut(&types.TxOutput{
			Amount:    uint64(tax),
			PkScript: params.OrganizationPkScript,
		})
	}
	// nulldata.
	if opReturnPkScript != nil {
		tx.AddTxOut(&types.TxOutput{
			Amount:    0,
			PkScript: opReturnPkScript,
		})
	}
	// AmountIn.
	//tx.TxIn[0].AmountIn = subsidy + uint64(tax)  //TODO, remove type conversion
	return types.NewTx(tx), nil
}

//calc coinbase
func (h *BlockHeader) CalcCoinBase(cfg *common.GlobalConfig,coinbaseStr string, extraNonce uint64,payAddressS string) error{
	transactions := make(Transactionses,0)
	if !h.HasCoinbasePack {
		h.TotalFee = 0
		for i:=0;i<len(h.Transactions);i++{
			transactions = append(transactions,h.Transactions[i])
		}
		sort.Sort(transactions)
		for i:=0;i<len(transactions);i++{
			h.TotalFee += transactions[i].Fee
		}
	}
	payToAddress,err := address.DecodeAddress(payAddressS)
	if err != nil {
		return err
	}
	coinbaseScript, err := standardCoinbaseScript(coinbaseStr,h.Height, extraNonce)
	if err != nil {
		return err
	}
	opReturnPkScript, err := standardCoinbaseOpReturn([]byte{})
	if err != nil {
		return err
	}
	//uit := 100000000
	coinbaseTx, err := createCoinbaseTx(uint64(h.Coinbasevalue)-h.TotalFee,
		coinbaseScript,
		opReturnPkScript,
		payToAddress,
		cfg.NecessaryConfig.Param)
	if err != nil{
		common.MinerLoger.Info(err.Error())
		return err
	}

	transactions = make(Transactionses,0)
	totalTxFee := uint64(0)
	if !h.HasCoinbasePack {
		tmpTrx := make(Transactionses,0)
		for i:=0;i<len(h.Transactions);i++{
			tmpTrx = append(tmpTrx,h.Transactions[i])
		}
		sort.Sort(tmpTrx)
		allSigCount := 0
		//every time pack max 1000 transactions and max 5000 sign scripts
		txCount := len(tmpTrx)
		if txCount>(cfg.OptionConfig.MaxTxCount - 1){
			txCount = cfg.OptionConfig.MaxTxCount - 1
		}
		for i:=0;i<txCount;i++{
			if allSigCount > (cfg.OptionConfig.MaxSigCount - 1){
				break
			}
			transactions = append(transactions,tmpTrx[i])
			allSigCount += tmpTrx[i].GetSigCount()
		}
		for i:=0;i<len(transactions);i++{
			totalTxFee += transactions[i].Fee
		}
	} else{
		for i:=1;i<len(h.Transactions);i++{
			totalTxFee += h.Transactions[i].Fee
		}
	}
	txBuf,err := coinbaseTx.Tx.Serialize()
	if err != nil {
		context := "Failed to serialize transaction"
		common.MinerLoger.Error(context)
		return err
	}
	// miner get tx tax
	coinbaseTx.Tx.TxOut[0].Amount += uint64(totalTxFee)
	if !h.HasCoinbasePack {
		newtransactions := make(Transactionses,0)
		newtransactions = append(newtransactions,Transactions{coinbaseTx.Tx.TxHash(),hex.EncodeToString(txBuf),0})
		newtransactions = append(newtransactions,transactions...)
		h.Transactions = newtransactions
		h.HasCoinbasePack = true
	} else {
		h.Transactions[0] = Transactions{coinbaseTx.Tx.TxHash(),hex.EncodeToString(txBuf),0}
	}
	return nil
}
