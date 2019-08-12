package qitmeer

import (
	"bytes"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	s "github.com/HalalChain/qitmeer-lib/core/serialization"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer-lib/core/types/pow"
	"io"
	"sync"
)

//qitmeer block header
type BlockHeader struct {
	sync.Mutex
	// block version
	Version uint32 `json:"version"`
	// The merkle root of the previous parent blocks (the dag layer)
	ParentRoot hash.Hash `json:"previousblockhash"`
	// The merkle root of the tx tree  (tx of the block)
	// included Witness here instead of the separated witness commitment
	TxRoot hash.Hash `json:"tx_root"`
	// The Multiset hash of UTXO set or(?) merkle range/path or(?) tire tree root
	// can all of the state data (stake, receipt, utxo) in state root?
	StateRoot hash.Hash `json:"stateroot"`

	// Difficulty target for tx
	Difficulty   uint32         `json:"difficulty"`
	Transactions []Transactions `json:"transactions"`
	Parents []ParentItems `json:"parents"`
	// Difficulty target for tx
	Bits string `json:"blake2bd_bits"`

	// block number
	Height uint64 `json:"height"`

	// TimeStamp
	Curtime uint32 `json:"curtime"`

	Pow pow.IPow

	// Nonce
	Target string `json:"blake2bd_target"`
	CuckarooTarget uint64 `json:"cuckaroo_target"`
	CuckatooTarget uint64 `json:"cuckatoo_target"`
	CuckarooScale uint64 `json:"cuckaroo_scale"`
	CuckatooScale uint64 `json:"cuckatoo_scale"`

	Coinbasevalue   int64 `json:"coinbasevalue"`
	HasCoinbasePack bool
}

//qitmeer block header
func BlockDataWithProof(h *types.BlockHeader) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, MaxBlockHeaderPayload))
	// TODO, redefine the protocol version and storage
	_ = writeBlockHeaderWithProof(buf, 0, h)
	return buf.Bytes()
}

func writeBlockHeaderWithProof(w io.Writer, pver uint32, bh *types.BlockHeader) error {
	sec := bh.Timestamp.Unix()
	return s.WriteElements(w, bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, bh.Difficulty, sec, bh.Pow)
}

// readBlockHeader reads a block header from io reader.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the type.
// TODO, redefine the protocol version and storage
func ReadBlockHeader(b []byte,bh *types.BlockHeader) error {
	r := bytes.NewReader(b)
	// TODO fix time ambiguous
	return s.ReadElements(r, &bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, &bh.Difficulty,(*s.Int64Time)(&bh.Timestamp),
		&bh.Pow)
}
