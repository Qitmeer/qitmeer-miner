package hlc

import "qitmeer/common/hash"

type ParentItems struct {
	Hash hash.Hash `json:"hash"`
	Data string    `json:"data"`
}

type Transactions struct {
	Hash hash.Hash `json:"hash"`
	Data string    `json:"data"`
	Fee  int64     `json:"fee"`
}
