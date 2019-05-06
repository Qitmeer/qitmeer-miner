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
