package qitmeer

import (
	"github.com/Qitmeer/qitmeer/core/types/pow"
)

const (
	//every mode position
	POWTYPE_START = 112
	POWTYPE_END   = 113
	TIMESTART     = 104
	TIMEEND       = 108
	NONCESTART    = 108
	NONCEEND      = 112
	HEIGHTSTART   = 104
	HEIGHTEND     = 112
	NBITSTART     = 100
	NBITEND       = 104
	STATESTART    = 68
	STATEEND      = 100
	MERKLESTART   = 36
	MERKLEEND     = 68
	PRESTART      = 4
	PREEND        = 36
	TXEND         = 68
	VERSIONSTART  = 0
	VERSIONEND    = 4
)

func CuckarooGraphWeight(mheight, targetHeight int64, edge_bits uint) uint64 {
	//45 days

	scale := (2 << (edge_bits - pow.MIN_CUCKAROOEDGEBITS)) * uint64(edge_bits)
	if scale <= 0 {
		scale = 1
	}
	return scale
}

func CuckatooGraphWeight(mheight, targetHeight int64, edge_bits uint) uint64 {
	//45 days
	return (2 << (edge_bits - pow.MIN_CUCKAROOEDGEBITS)) * uint64(edge_bits)
}
