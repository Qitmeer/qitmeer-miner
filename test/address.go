package test

import (
	"github.com/noxproject/nox/crypto/seed"
	"fmt"
	"log"
	"github.com/noxproject/nox/crypto/bip32"
	"encoding/hex"
	"github.com/noxproject/nox/crypto/ecc"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/common/encode/base58"
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
	switch curve {
	case "secp256k1":
		fmt.Println(len(entropy))
		masterKey,err := bip32.NewMasterKey(entropy)
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

