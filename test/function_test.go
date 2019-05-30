package test

import (
	"hlc-miner/common/qitmeer/address"
	"hlc-miner/common/qitmeer/params"
	"testing"
	"log"
	"hlc-miner/common"
)

func TestCreateNoxAddr(t *testing.T) {
	seed := newEntropy(32)
	log.Println("【rand seed】",seed)
	privateKey := ecNew("ed25519",seed)
	log.Println("【private key】",privateKey)
	publicKey := ecPrivateKeyToEcPublicKey(false,privateKey)
	log.Println("【public key】",publicKey)
	//param := params.PrivNetParams
	param := params.TestNetParams
	//param := params.MainNetParams
	addr := ecPubKeyToAddress(param.PubKeyHashAddrID[:],publicKey)
	addres,err := address.DecodeAddress(addr)
	if err != nil{
		log.Fatalln("【验证地址失败】",err)
		return
	}
	log.Println("【base58 address】",addres)
}

func TestRand(t *testing.T)  {
	a := common.GenerateRand(32)

	log.Println(a)
}
