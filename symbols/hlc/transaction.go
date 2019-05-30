package hlc

import "hlc-miner/common/qitmeer/hash"

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
