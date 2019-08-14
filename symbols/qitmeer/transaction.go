package qitmeer

import (
	"bytes"
	"encoding/hex"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/core/message"
	"github.com/HalalChain/qitmeer-lib/core/protocol"
)

const MAX_SIG_COUNT = 5000
const MAX_TX_COUNT = 1000
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


func (this *Transactions) GetSigCount() int{
	txBytes,_ := hex.DecodeString(this.Data)
	var mtx *message.MsgTx
	mtx = new(message.MsgTx)
	_ = mtx.Decode(bytes.NewReader(txBytes),protocol.ProtocolVersion)
	return len(mtx.Tx.TxOut)
}
