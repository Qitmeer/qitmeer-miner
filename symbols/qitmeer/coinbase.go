// Copyright (c) 2019 The halalchain developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qitmeer

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/core/address"
	s "github.com/HalalChain/qitmeer-lib/core/serialization"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer-lib/engine/txscript"
	"github.com/HalalChain/qitmeer-lib/params"
	"log"
	"qitmeer-miner/common"
	"sort"
)

func standardCoinbaseOpReturn(height uint32, extraNonce uint64) ([]byte, error) {
	enData := make([]byte, 12)
	binary.LittleEndian.PutUint32(enData[0:4], height)
	binary.LittleEndian.PutUint64(enData[4:12], extraNonce)
	extraNonceScript, err := txscript.GenerateProvablyPruneableOut(enData)
	if err != nil {
		return nil, err
	}

	return extraNonceScript, nil
}

func qitmeerCoinBase(coinbaseVal int, coinbaseScript []byte, opReturnPkScript []byte, nextBlockHeight int64,
	addr types.Address, voters uint16, params *params.Params)(*types.Tx, error){
	tx := types.NewTransaction()
	tx.AddTxIn(&types.TxInput{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOut: *types.NewOutPoint(&hash.Hash{},
			types.MaxPrevOutIndex),
		Sequence:    types.MaxTxInSequenceNum,
		BlockOrder: types.NullBlockOrder,
		TxIndex:     types.NullTxIndex,
		SignScript:  coinbaseScript,
	})

	// Block one is a special block that might pay out tokens to a ledger.
	if nextBlockHeight == 1 && len(params.BlockOneLedger) != 0 {
		// Convert the addresses in the ledger into useable format.
		addrs := make([]types.Address, len(params.BlockOneLedger))
		for i, payout := range params.BlockOneLedger {
			addr, err := address.DecodeAddress(payout.Address)
			if err != nil {
				return nil, err
			}
			addrs[i] = addr
		}

		for i, payout := range params.BlockOneLedger {
			// Make payout to this address.
			pks, err := txscript.PayToAddrScript(addrs[i])
			if err != nil {
				return nil, err
			}
			tx.AddTxOut(&types.TxOutput{
				Amount:   payout.Amount,
				PkScript: pks,
			})
		}
		tx.TxIn[0].AmountIn = params.BlockOneSubsidy()

		return types.NewTx(tx), nil
	}
	// Create a coinbase with correct block subsidy and extranonce.
	//subsidy := uint64(coinbaseVal)
	allRate := params.BlockTaxProportion + params.WorkRewardProportion
	subsidy := float64(coinbaseVal) * float64(params.WorkRewardProportion) / float64(allRate)
	tax := float64(coinbaseVal) * float64(params.BlockTaxProportion) / float64(allRate)
	// Tax output.
	if params.BlockTaxProportion > 0 {
		tx.AddTxOut(&types.TxOutput{
			Amount:   uint64(tax),
			PkScript: params.OrganizationPkScript,
		})
	} else {
		// Tax disabled.
		scriptBuilder := txscript.NewScriptBuilder()
		trueScript, err := scriptBuilder.AddOp(txscript.OP_TRUE).Script()
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(&types.TxOutput{
			Amount:   uint64(tax),
			PkScript: trueScript,
		})
	}
	// Extranonce.
	tx.AddTxOut(&types.TxOutput{
		Amount:   0,
		PkScript: opReturnPkScript,
	})
	// AmountIn.
	tx.TxIn[0].AmountIn = uint64(subsidy) + uint64(tax) //TODO, remove type conversion

	// Create the script to pay to the provided payment address if one was
	// specified.  Otherwise create a script that allows the coinbase to be
	// redeemable by anyone.
	var pksSubsidy []byte
	if addr != nil {
		var err error
		pksSubsidy, err = txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		scriptBuilder := txscript.NewScriptBuilder()
		pksSubsidy, err = scriptBuilder.AddOp(txscript.OP_TRUE).Script()
		if err != nil {
			return nil, err
		}
	}
	// Subsidy paid to miner.
	tx.AddTxOut(&types.TxOutput{
		Amount:   uint64(subsidy),
		PkScript: pksSubsidy,
	})
	return types.NewTx(tx), nil
}

//calc coinbase
func (h *BlockHeader) CalcCoinBase(coinbaseStr string,payAddress string) error{
	coinbaseScript := []byte{0x00, 0x00}
	coinbaseScript = append(coinbaseScript, []byte(coinbaseStr)...)
	rand, err := s.RandomUint64()
	if err != nil {
		log.Println(err)
		return err
	}
	nextBlockHeight := uint32(h.Height)
	opReturnPkScript, err := standardCoinbaseOpReturn(nextBlockHeight,
		rand)
	if err != nil {
		log.Println(err)
		return err
	}
	payToAddress,err := address.DecodeAddress(payAddress)
	if err != nil {
		log.Println(err)
		return err
	}
	voters := 0 //TODO remove voters
	params1 := &params.Params{}
	//miner
	params1.WorkRewardProportion = 9
	//stake
	params1.StakeRewardProportion = 0
	//team
	params1.BlockTaxProportion = 1
	//group
	params1.OrganizationPkScript = common.HexMustDecode("76a914699e7e705893b4e7b3f9742ca55a743c7167288a88ac")
	coinbaseTx, err := qitmeerCoinBase(int(h.Coinbasevalue),
		coinbaseScript,
		opReturnPkScript,
		int64(nextBlockHeight), //TODO remove type conversion
		payToAddress,
		uint16(voters),
		params1)
	if err != nil{
		log.Println(err)
		return err
	}
	transactions := make(Transactionses,0)
	totalTxFee := int64(0)
	if !h.HasCoinbasePack {
		tmpTrx := make(Transactionses,0)
		for i:=0;i<len(h.Transactions);i++{
			tmpTrx = append(tmpTrx,h.Transactions[i])
		}
		sort.Sort(tmpTrx)
		allSigCount := 0
		//every time pack max 1000 transactions and max 5000 sign scripts
		txCount := len(tmpTrx)
		if txCount>MAX_TX_COUNT{
			txCount = MAX_TX_COUNT
		}
		for i:=0;i<txCount;i++{
			if allSigCount > MAX_SIG_COUNT - 1{
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
	coinbaseTx.Tx.TxOut[2].Amount += uint64(totalTxFee)
	txBuf,err := coinbaseTx.Tx.Serialize(types.TxSerializeFull)
	if err != nil {
		context := "Failed to serialize transaction"
		log.Println(context)
		return err
	}
	if !h.HasCoinbasePack {
		newtransactions := make(Transactionses,0)
		newtransactions = append(newtransactions,Transactions{coinbaseTx.Tx.TxHashFull(),hex.EncodeToString(txBuf),0})
		newtransactions = append(newtransactions,transactions...)
		h.Transactions = newtransactions
		h.HasCoinbasePack = true
	} else {
		h.Transactions[0] = Transactions{coinbaseTx.Tx.TxHashFull(),hex.EncodeToString(txBuf),0}
	}
	return nil
}
