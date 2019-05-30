package test

import (
	"fmt"
	"hlc-miner/common/qitmeer/base58"
	"hlc-miner/common/qitmeer/bip32"
	"hlc-miner/common/qitmeer/ecc"
	"hlc-miner/common/qitmeer/hash"
	"hlc-miner/common/qitmeer/seed"
	"log"
	"encoding/hex"
)
//generate seed
func newEntropy(size uint) string{
	s,err :=seed.GenerateSeed(uint16(size))
	if err!=nil {
		log.Fatal(err)
		return ""
	}
	return fmt.Sprintf("%x",s)
}
//secp256k1 generate private key
func ecNew(curve string, entropyStr string) string{
	entropy, err := hex.DecodeString(entropyStr)
	if err!=nil {
		log.Fatalln("【error】",entropyStr,err)
		return ""
	}
	masterKey := &bip32.Key{}
	switch curve {
	case "secp256k1":
		masterKey,err = bip32.NewMasterKey(entropy)
		if err!=nil {
			log.Fatalln(err)
			return ""
		}
	case "ed25519":
		masterKey,err = bip32.NewMasterKey(entropy)
		if err!=nil {
			log.Fatalln(err)
			return ""
		}
		return fmt.Sprintf("%x",masterKey.Key[:])
	default:
	}
	return ""
}

//from private key to public key
func ecPrivateKeyToEcPublicKey(uncompressed bool, privateKeyStr string) string{
	data, err := hex.DecodeString(privateKeyStr)
	if err!=nil {
		log.Fatalln(err)
		return ""
	}
	_, pubKey := ecc.Secp256k1.PrivKeyFromBytes(data)
	var key []byte
	if uncompressed {
		key = pubKey.SerializeUncompressed()
	}else {
		key = pubKey.SerializeCompressed()
	}
	return fmt.Sprintf("%x",key[:])
}

// public key to bas58 address
func ecPubKeyToAddress(version []byte, pubkey string) string{
	data, err :=hex.DecodeString(pubkey)
	if err != nil {
		log.Println(err)
		return ""
	}
	h := hash.Hash160(data)

	address := base58.NoxCheckEncode(h, version[:])
	return fmt.Sprintf("%s",address)
}

