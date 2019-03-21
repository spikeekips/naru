package common

import (
	"strconv"

	sebakcommon "boscoin.io/sebak/lib/common"
)

type Amount uint64

func NewFromSEBAKAmount(a sebakcommon.Amount) Amount {
	return Amount(uint64(a))
}

func (a Amount) String() string {
	return strconv.FormatUint(uint64(a), 10)
}
