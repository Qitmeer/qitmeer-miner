package qitmeer

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer-lib/core/types/pow"
	"math/big"
	"qitmeer-miner/common"
	"time"
)

type MinerBlockData struct {
	Transactions []Transactions
	Parents []ParentItems
	HeaderData []byte
	TargetDiff *big.Int
	JobID string
	HeaderBlock *types.BlockHeader
}
// Header structure of assembly pool
func BlockComputePoolData(b []byte) []byte{
	//the qitmeer order
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
func (this *MinerBlockData)PackagePoolHeader(work *QitmeerWork,powType pow.PowType)  {
	this.HeaderData = BlockComputePoolData(work.PoolWork.WorkData) // 128
	this.TargetDiff = work.stra.Target
	nbitesBy := common.Target2BlockBits(fmt.Sprintf("%064x",this.TargetDiff))
	copy(this.HeaderData[NONCESTART:NONCEEND],nbitesBy[:])
	instance := pow.GetInstance(powType,0,[]byte{})
	proofData,_ := hex.DecodeString(instance.GetProofData())
	this.HeaderData = append(this.HeaderData,proofData...) //328 bytes
	this.JobID = work.PoolWork.JobID
	this.HeaderBlock = &types.BlockHeader{}
	_ = ReadBlockHeader(this.HeaderData,this.HeaderBlock)
}
//the pool work submit structure
func (this *MinerBlockData)PackagePoolHeaderByNonce(work *QitmeerWork,nonce uint64)  {
	this.HeaderData = BlockComputePoolData(work.PoolWork.WorkData)
	this.TargetDiff = work.stra.Target
	nbitesBy := make([]byte,8)
	binary.LittleEndian.PutUint64(nbitesBy,nonce)
	copy(this.HeaderData[NONCESTART:NONCEEND],nbitesBy[:])
	this.JobID = work.PoolWork.JobID
}

//the solo work submit structure
func (this *MinerBlockData)PackageRpcHeader(work *QitmeerWork)  {
	bitesBy ,_:= hex.DecodeString(work.Block.Target)
	bitesBy = common.Reverse(bitesBy[:8])
	this.Parents = work.Block.Parents
	this.Transactions = make([]Transactions,0)
	for i:=0;i<len(work.Block.Transactions);i++{
		this.Transactions = append(this.Transactions,Transactions{
			work.Block.Transactions[i].Hash,work.Block.Transactions[i].Data,work.Block.Transactions[i].Fee,
		})
	}
	b1 , _ := hex.DecodeString(work.Block.Target)
	var r [32]byte
	copy(r[:],common.Reverse(b1)[:])
	r1 := hash.Hash(r)
	this.TargetDiff = HashToBig(&r1)
	this.HeaderBlock = &types.BlockHeader{}
	this.HeaderBlock.Version = work.Block.Version
	this.HeaderBlock.ParentRoot = work.Block.ParentRoot
	this.HeaderBlock.TxRoot = work.Block.TxRoot
	this.HeaderBlock.StateRoot = work.Block.StateRoot
	this.HeaderBlock.Difficulty = work.Block.Difficulty
	this.HeaderBlock.Timestamp = time.Unix(int64(work.Block.Curtime), 0)
	this.HeaderBlock.Pow = pow.GetInstance(work.Block.Pow.GetPowType(),binary.LittleEndian.Uint64(bitesBy),[]byte{})

}