package cuckoo

import (
	"encoding/binary"
	"errors"
	"log"
)
type sip struct {
	k0 uint64
	k1 uint64
	v  [4]uint64
}
func newsip(h []byte) *sip {
	s := &sip{
		k0: binary.LittleEndian.Uint64(h[:]),
		k1: binary.LittleEndian.Uint64(h[8:]),
	}
	s.v[0] = s.k0 ^ 0x736f6d6570736575
	s.v[1] = s.k1 ^ 0x646f72616e646f6d
	s.v[2] = s.k0 ^ 0x6c7967656e657261
	s.v[3] = s.k1 ^ 0x7465646279746573
	return s
}

//Verify verifiex cockoo nonces.
func Verify(sipkey []byte, nonces []uint32) error {
	sip := newsip(sipkey)
	var uvs [2 * PROOF_SIZE]uint32
	var xor0, xor1 uint32

	if len(nonces) != PROOF_SIZE {
		return errors.New("length of nonce is not correct")
	}

	if nonces[PROOF_SIZE-1] > easiness {
		return errors.New("nonce is too big")
	}

	for n := 0; n < PROOF_SIZE; n++ {
		if n > 0 && nonces[n] <= nonces[n-1] {
			return errors.New("nonces are not in order")
		}
		u00 := siphashPRF(&sip.v, uint64(nonces[n]<<1))
		v00 := siphashPRF(&sip.v, (uint64(nonces[n])<<1)|1)
		u0 := uint32(u00&edgemask) << 1
		xor0 ^= u0
		uvs[2*n] = u0
		v0 := (uint32((v00)&edgemask) << 1) | 1
		xor1 ^= v0
		uvs[2*n+1] = v0
	}


	if xor0 != 0 {
		return errors.New("U endpoinsts don't match")
	}
	if xor1 != 0 {
		return errors.New("V endpoinsts don't match")
	}
	n := 0
	log.Println(uvs)
	for i := 0; ; {
		another := i
		for k := (i + 2) % (2 * PROOF_SIZE); k != i; k = (k + 2) % (2 * PROOF_SIZE) {

			if uvs[k] == uvs[i] {
				if another != i {
					return errors.New("there are branches in nonce")
				}
				another = k
			}
		}
		if another == i {
			return errors.New("dead end in nonce")
		}
		i = another ^ 1
		n++
		if i == 0 {
			break
		}
	}
	if n != PROOF_SIZE {
		return errors.New("cycle is too short")
	}
	return nil
}
