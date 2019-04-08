package common

import (
	"fmt"
	"strconv"
	"strings"

	sebakcommon "boscoin.io/sebak/lib/common"
)

type Amount uint64

func NewFromSEBAKAmount(a sebakcommon.Amount) Amount {
	return Amount(uint64(a))
}

func (a Amount) String() string {
	return strconv.FormatUint(uint64(a), 10)
}

func GonToBOS(i interface{}) string {
	var s string
	switch i.(type) {
	case uint64:
		s = strconv.FormatUint(i.(uint64), 10)
	case Amount:
		s = i.(Amount).String()
	case sebakcommon.Amount:
		s = i.(sebakcommon.Amount).String()
	}

	if len(s) < 7 {
		return fmt.Sprintf(
			"0.%s%s",
			strings.Repeat("0", 7-len(s)),
			s,
		)
	}

	if len(s) == 7 {
		return fmt.Sprintf("0.%s", s)
	}

	return fmt.Sprintf("%s.%s", s[:len(s)-7], s[len(s)-7:])
}
