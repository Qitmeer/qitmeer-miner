package qitmeer

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
)

type ParentItems struct {
	Hash hash.Hash `json:"hash"`
	Data string    `json:"data"`
}

type Transactions struct {
	Hash hash.Hash `json:"hash"`
	Data string    `json:"data"`
	Fee  int64     `json:"fee"`
}

type Transactionses []Transactions

func (p Transactionses) Len() int { return len(p) }

// fee sort desc
func (p Transactionses) Less(i, j int) bool {
	return p[i].Fee > p[j].Fee
}
func (p Transactionses) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (this *Transactions) GetSigCount() int {
	txBytes, _ := hex.DecodeString(this.Data)
	mtx := types.NewTransaction()
	_ = mtx.Decode(bytes.NewReader(txBytes), protocol.ProtocolVersion)
	return len(mtx.TxOut)
}

func (this *Transactions) EncodeTx() *types.Tx {
	txBytes, _ := hex.DecodeString(this.Data)
	mtx := types.NewTransaction()
	_ = mtx.Decode(bytes.NewReader(txBytes), protocol.ProtocolVersion)
	_ = mtx.Decode(bytes.NewReader(txBytes), protocol.ProtocolVersion)
	tx := types.NewTx(mtx)
	return tx
}
