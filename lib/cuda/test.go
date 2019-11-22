package main

import (
	`encoding/binary`
	`encoding/hex`
	`fmt`
	`github.com/Qitmeer/qitmeer/common/hash`
	`github.com/Qitmeer/qitmeer/crypto/cuckoo`
	`log`
	`unsafe`
)

//#cgo CFLAGS: -I.
//#cgo LDFLAGS: -L. -lcudacuckoo
//#cgo LDFLAGS: -lcudart
//#include "test.h"
//#include <stdio.h>
//#include <stdlib.h>
import "C"


func main() {
	fmt.Printf("Invoking cuda library...\n")
		header,_ := hex.DecodeString("09000000c188c819f82ca290231c8d3f67bff511c0918672fe228d63c7654a11390ca101cde8702772b101c7dfff012b1fc87e60a2be5136dd062349401a0d077f9ef428000000000000000000000000000000000000000000000000000000000000000000c60902900000000000000001")
		deviceID := 0
		cycleNoncesBytes := make([]byte,42*4)
		nonceBytes := make([]byte,4)
		resultBytes := make([]byte,4)
		fmt.Println(len(header),hex.EncodeToString(header))
		_ = C.cuda_search((C.int)(deviceID),(*C.uchar)(unsafe.Pointer(&header[0])),(*C.uint)(unsafe.Pointer(&resultBytes[0])),(*C.uint)(unsafe.Pointer(&nonceBytes[0])),(*C.uint)(unsafe.Pointer(&cycleNoncesBytes[0])))

		nonces := make([]uint32,0)
		copy(header[108:112],nonceBytes)

		for jj := 0;jj < len(cycleNoncesBytes);jj+=4{
			tj := binary.LittleEndian.Uint32(cycleNoncesBytes[jj:jj+4])
			if tj <=0 {
				break
			}
			nonces = append(nonces,tj)
		}
		h := hash.HashH(header)
		fmt.Println(h)
		fmt.Println(nonces)
		err := cuckoo.VerifyCuckaroo(h[:], nonces, 24)

		if err != nil {
			log.Println("verify error", err)
			return
		}

		log.Println("verify success")

}