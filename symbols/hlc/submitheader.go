package hlc

import (
	"encoding/hex"
	"fmt"
	"hlc-miner/common"
	"hlc-miner/common/qitmeer/blockchain"
	"hlc-miner/common/qitmeer/hash"
	"math/big"
)

type MinerBlockData struct {
	Transactions []Transactions
	Parents []ParentItems
	HeaderData []byte
	TargetDiff *big.Int
	JobID string
}
// Header structure of assembly pool
func BlockComputePoolData(b []byte) []byte{
	//the nox order
	nonce := hex.EncodeToString(b[NONCESTART:NONCEEND])
	ntime := hex.EncodeToString(b[TIMESTART:TIMEEND])
	height := hex.EncodeToString(b[HEIGHTSTART:HEIGHTEND])
	nbits := hex.EncodeToString(b[NBITSTART:NBITEND])
	state := hex.EncodeToString(b[STATESTART:STATEEND])
	merkle := hex.EncodeToString(b[MERKLESTART:MERKLEEND])
	prevhash := hex.EncodeToString(b[PRESTART:PREEND])
	version := hex.EncodeToString(b[VERSIONSTART:VERSIONEND])
	//the pool order
	header := nonce +ntime+height+ nbits + state + merkle + prevhash + version

	bb , _ := hex.DecodeString(header)
	bb = common.Reverse(bb)
	heightD := common.Reverse(bb[HEIGHTSTART:HEIGHTEND])
	copy(bb[HEIGHTSTART:HEIGHTEND],heightD[0:8])
	return bb
}
//the pool work submit structure
func (this *MinerBlockData)PackagePoolHeader(work *HLCWork)  {
	this.HeaderData = BlockComputePoolData(work.PoolWork.WorkData)
	this.TargetDiff = work.stra.Target
	nbitesBy := common.Target2BlockBits(fmt.Sprintf("%064x",this.TargetDiff))
	copy(this.HeaderData[NONCESTART:NONCEEND],nbitesBy[:])
	this.JobID = work.PoolWork.JobID
}

//the solo work submit structure
func (this *MinerBlockData)PackageRpcHeader(work *HLCWork)  {
	//log.Println(work.Block.Target)
	bitesBy ,_:= hex.DecodeString(work.Block.Target)
	bitesBy = common.Reverse(bitesBy[:8])
	this.HeaderData = work.Block.BlockData()
	this.Transactions = work.Block.Transactions
	this.Parents = work.Block.Parents
	copy(this.HeaderData[NONCESTART:NONCEEND],bitesBy[:])

	b1 , _ := hex.DecodeString(work.Block.Target)
	var r [32]byte
	copy(r[:],common.Reverse(b1)[:])
	r1 := hash.Hash(r)
	this.TargetDiff = blockchain.HashToBig(&r1)
}
