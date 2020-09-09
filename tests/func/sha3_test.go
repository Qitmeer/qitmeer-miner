package _func

import (
	"ctkcontract/crypto"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"testing"
)

func TestSha3(t *testing.T) {
	b := []byte("helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworld")
	fmt.Println(hex.EncodeToString(b[:]))
	h := crypto.Keccak256(b)
	fmt.Println(hex.EncodeToString(h[:]))
	h1 := hash.HashQitmeerKeccak256(b)
	fmt.Println(hex.EncodeToString(h1[:]))
}

func TestSha3Q(t *testing.T) {
	b := []byte("helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhel")
	fmt.Println(hex.EncodeToString(b[:]))
	h := crypto.Keccak256(b)
	fmt.Println(hex.EncodeToString(h[:]))
	h1 := hash.HashQitmeerKeccak256(b)
	fmt.Println(hex.EncodeToString(h1[:]))
	fmt.Println(h1)
}
