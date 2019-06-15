package cuckoo


func unsignedShift(val uint, amt uint) uint {
	var mask uint
	mask = (1 << (64 - amt)) - 1
	return (val >> amt) & mask
}

var V [4]uint
func Siphash24(nonce uint) uint {
	var v0, v1, v2, v3 uint
	v0 = V[0]
	v1 = V[1]
	v2 = V[2]
	v3 = V[3] ^ nonce

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | unsignedShift(v1, 51)
	v3 = (v3 << 16) | unsignedShift(v3, 48)
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | unsignedShift(v0, 32)
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | unsignedShift(v1, 47)
	v3 = (v3 << 21) | unsignedShift(v3, 43)
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | unsignedShift(v2, 32)

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | unsignedShift(v1, 51)
	v3 = (v3 << 16) | unsignedShift(v3, 48)
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | unsignedShift(v0, 32)
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | unsignedShift(v1, 47)
	v3 = (v3 << 21) | unsignedShift(v3, 43)
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | unsignedShift(v2, 32)

	v0 ^= nonce
	v2 ^= 0xff

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | unsignedShift(v1, 51)
	v3 = (v3 << 16) | unsignedShift(v3, 48)
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | unsignedShift(v0, 32)
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | unsignedShift(v1, 47)
	v3 = (v3 << 21) | unsignedShift(v3, 43)
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | unsignedShift(v2, 32)

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | unsignedShift(v1, 51)
	v3 = (v3 << 16) | unsignedShift(v3, 48)
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | unsignedShift(v0, 32)
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | unsignedShift(v1, 47)
	v3 = (v3 << 21) | unsignedShift(v3, 43)
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | unsignedShift(v2, 32)

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | unsignedShift(v1, 51)
	v3 = (v3 << 16) | unsignedShift(v3, 48)
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | unsignedShift(v0, 32)
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | unsignedShift(v1, 47)
	v3 = (v3 << 21) | unsignedShift(v3, 43)
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | unsignedShift(v2, 32)

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | unsignedShift(v1, 51)
	v3 = (v3 << 16) | unsignedShift(v3, 48)
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | unsignedShift(v0, 32)
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | unsignedShift(v1, 47)
	v3 = (v3 << 21) | unsignedShift(v3, 43)
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | unsignedShift(v2, 32)

	return v0 ^ v1 ^ v2 ^ v3
}


func siphashPRF(v *[4]uint64, b uint64) uint64 {
	v0 := v[0]
	v1 := v[1]
	v2 := v[2]
	v3 := v[3]
	// Initialization.
	// Compression.
	v3 ^= b

	// Round 1.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 2.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	v0 ^= b

	// Finalization.
	v2 ^= 0xff

	// Round 1.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 2.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 3.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 4.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	return v0 ^ v1 ^ v2 ^ v3
}