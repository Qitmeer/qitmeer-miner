// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qitmeer

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/symbols/qitmeer/coinbase"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/merkle"
	"github.com/Qitmeer/qitmeer/core/types"
)

//calc coinbase
func (h *BlockHeader) CalcCoinBase(cfg *common.GlobalConfig, coinbaseStr string, extraNonce uint64, payAddressS string) (*hash.Hash, []Transactions) {
	transactions := make(Transactionses, 0)
	txs := make([]*types.Tx, 0)
	if !h.HasCoinbasePack {
		h.TotalFee = 0
		for i := 0; i < len(h.Transactions); i++ {
			transactions = append(transactions, h.Transactions[i])
			txs = append(txs, h.Transactions[i].EncodeTx())
			h.transactions = append(h.transactions, h.Transactions[i].EncodeTx())
			h.TotalFee += h.Transactions[i].Fee
		}
	}
	instance := coinbase.GetNewCoinbaseInstance(int(h.Version), cfg.NecessaryConfig.Param, payAddressS,
		coinbaseStr, extraNonce, h.Height, h.TotalFee, uint64(h.Coinbasevalue), h.TotalFee)
	// miner get tx tax
	coinbaseTx, opPkReturnScript := instance.GetCoinbaseTx()
	if coinbaseTx == nil {
		return nil, []Transactions{}
	}
	blockFeesMap := types.AmountMap{}
	for cid, val := range h.BlockFeesMap {
		blockFeesMap[types.CoinID(cid)] = val
	}
	err := fillOutputsToCoinBase(coinbaseTx, blockFeesMap, nil, opPkReturnScript)
	if err != nil {
		context := "Failed to fillOutputsToCoinBase"
		common.MinerLoger.Error(context)
		return nil, []Transactions{}
	}
	if !h.HasCoinbasePack {
		newtransactions := make(Transactionses, 0)
		newtransactions = append(newtransactions, Transactions{coinbaseTx.Tx.TxHash(), "", 0})
		newtransactions = append(newtransactions, transactions...)
		transactions = newtransactions
		h.Transactions = transactions
		h.HasCoinbasePack = true
		ntxs := make([]*types.Tx, 0)
		ntxs = append(ntxs, coinbaseTx)
		ntxs = append(ntxs, txs...)
		txs = ntxs
		h.transactions = ntxs
	} else {
		transactions = h.Transactions
		txs = h.transactions
		transactions[0] = Transactions{coinbaseTx.Tx.TxHash(), "", 0}
		txs[0] = coinbaseTx
	}
	_ = fillWitnessToCoinBase(txs)
	txBuf, err := txs[0].Tx.Serialize()
	if err != nil {
		context := "Failed to serialize transaction"
		common.MinerLoger.Error(context)
		return nil, []Transactions{}
	}
	transactions[0].Hash = *(txs[0].Hash())
	transactions[0].Data = fmt.Sprintf("%x", txBuf)
	merkles := merkle.BuildMerkleTreeStore(txs, false)
	return merkles[len(merkles)-1], transactions
}

func (h *BlockHeader) AddCoinbaseTx(coinbaseTx *types.Tx) {
	if h.HasCoinbasePack {
		h.transactions[0] = coinbaseTx
	} else {
		txs := make([]*types.Tx, 0)
		txs = append(txs, coinbaseTx)
		txs = append(txs, h.transactions...)
		h.transactions = txs
	}
}

func fillWitnessToCoinBase(blockTxns []*types.Tx) error {
	merkles := merkle.BuildMerkleTreeStore(blockTxns, true)
	txWitnessRoot := merkles[len(merkles)-1]
	witnessPreimage := append(txWitnessRoot.Bytes(), blockTxns[0].Tx.TxIn[0].SignScript...)
	witnessCommitment := hash.DoubleHashH(witnessPreimage[:])
	blockTxns[0].Tx.TxIn[0].PreviousOut.Hash = witnessCommitment
	blockTxns[0].RefreshHash()
	return nil
}

func fillOutputsToCoinBase(coinbaseTx *types.Tx, blockFeesMap types.AmountMap,
	taxOutput *types.TxOutput, oprOutput *types.TxOutput) error {
	if len(coinbaseTx.Tx.TxOut) != blockchain.CoinbaseOutput_subsidy+1 {
		return fmt.Errorf("coinbase output error")
	}
	for k, v := range blockFeesMap {
		if v <= 0 || k == types.MEERID {
			continue
		}
		coinbaseTx.Tx.AddTxOut(&types.TxOutput{
			Amount:   types.Amount{Value: 0, Id: k},
			PkScript: coinbaseTx.Tx.TxOut[0].GetPkScript(),
		})
	}
	if taxOutput != nil {
		coinbaseTx.Tx.AddTxOut(taxOutput)
	}
	if oprOutput != nil {
		coinbaseTx.Tx.AddTxOut(oprOutput)
	}
	return nil
}
