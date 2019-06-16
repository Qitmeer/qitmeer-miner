package cuckoo

func CheckDiff(nonces []uint32)  bool {
	//packedSolution := make([]byte,0)
	//for i :=0;i<len(nonces);i++{
	//	b := make([]byte,4)
	//	binary.LittleEndian.PutUint32(b,nonces[i])
	//	packedSolution = append(packedSolution,b...)
	//}
	//hash1 := blake2b.Sum256(packedSolution)
	//
	//
	//h := hash.Hash(hash1)
	//b1 , _ := hex.DecodeString("0fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	//var r [32]byte
	//copy(r[:],common.Reverse(b1)[:])
	//r1 := hash.Hash(r)
	//log.Println(fmt.Sprintf("target:"),h)
	//targetDiff := blockchain.HashToBig(&r1)
	//if blockchain.HashToBig(&h).Cmp(targetDiff) <= 0 {
	//	fmt.Println("Found Solve")
	//	return true
	//}
	return false
}