package main

/*
#include "./meer/main.h"
#include "./meer/main.c"
#include "./meer/algo_meer.c"
#include "./meer/meer.h"
#include "./meer/meer_drv.c"
#include "./meer/meer_drv.h"
#include "./meer/uart.h"
#include "./meer/uart.c"
#cgo CFLAGS: -Wno-unused-result
#cgo CFLAGS: -Wno-int-conversion
*/
import "C"
import (
	"encoding/hex"
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("test")
	targetBytes, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000ffff00000000")
	headerBytes, _ := hex.DecodeString("1200000003c60b43da920ae08be3dd91e174fc7b5d538ca5601a4ea9fbcfc703447dd4871b7fac4e54a887df6c1801f4ac37883d6808cb93855f1f07aa4c2cfa73eea3b1000000000000000000000000000000000000000000000000000000000000000000f5231c83cf1060080000000000000000")
	end := []byte{0}
	nonce := make([]byte, 8)
	C.meer_pow((*C.char)(unsafe.Pointer(&headerBytes[0])), (C.int)(len(headerBytes)),
		(*C.char)(unsafe.Pointer(&targetBytes[0])),
		(*C.uchar)(unsafe.Pointer(&nonce[0])),
		(*C.uchar)(unsafe.Pointer(&end[0])),
	)
}
