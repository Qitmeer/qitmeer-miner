package qitmeer

import (
	"bytes"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	s "github.com/Qitmeer/qitmeer-lib/core/serialization"
	"github.com/Qitmeer/qitmeer-lib/core/types"
	"io"
)

//qitmeer block header
type BlockHeader struct {
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
	Bits string `json:"bits"`

	// block number
	Height uint64 `json:"height"`

	// TimeStamp
	Curtime uint32 `json:"curtime"`

	// Nonce
	Nonce uint64 `json:"nonce"`
	Nonces []*uint32 `json:"nonces"`
	Target string `json:"target"`

	Coinbasevalue   int64 `json:"coinbasevalue"`
	HasCoinbasePack bool
	TotalFee uint64
	transactions []*types.Tx
}

//qitmeer block header
func (h *BlockHeader) BlockData() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, MaxBlockHeaderPayload))
	// TODO, redefine the protocol version and storage
	_ = writeBlockHeader(buf, 0, h)
	return buf.Bytes()
}

//qitmeer block header
func BlockData(h *types.BlockHeader) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	// TODO, redefine the protocol version and storage
	_ = WriteBlockHeader(buf, h)
	return buf.Bytes()
}

//qitmeer block header
func (h *BlockHeader) BlockDataCuckaroo() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, MaxBlockHeaderPayload))
	// TODO, redefine the protocol version and storage
	_ = writeBlockHeaderCuckaroo(buf, 0, h)
	return buf.Bytes()
}

//qitmeer Header structure of assembly
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	sec := uint64(bh.Curtime)
	return s.WriteElements(w, bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, bh.Difficulty, bh.Height, sec, bh.Nonce)
}

//qitmeer Header structure of assembly
func WriteBlockHeader(w io.Writer, bh *types.BlockHeader) error {
	sec := bh.Timestamp.Unix()
	return s.WriteElements(w, bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, bh.Difficulty, bh.ExNonce, sec, bh.Nonce)
}

func writeBlockHeaderCuckaroo(w io.Writer, pver uint32, bh *BlockHeader) error {
	sec := uint64(bh.Curtime)
	return s.WriteElements(w, bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, bh.Difficulty, bh.Height, sec, bh.Nonce,bh.Nonces)
}

//block hash
func (h *BlockHeader) BlockHash() hash.Hash {
	buf := bytes.NewBuffer(make([]byte, 0, MaxBlockHeaderPayload))
	// TODO, redefine the protocol version and storage
	_ = writeBlockHeader(buf, 0, h)
	return hash.DoubleHashH(buf.Bytes())
}

// readBlockHeader reads a block header from io reader.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the type.
// TODO, redefine the protocol version and storage
func ReadBlockHeader(b []byte,bh *types.BlockHeader) error {
	r := bytes.NewReader(b)
	// TODO fix time ambiguous
	return s.ReadElements(r, &bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, &bh.Difficulty,&bh.ExNonce,(*s.Int64Time)(&bh.Timestamp),
		&bh.Nonce)
}


