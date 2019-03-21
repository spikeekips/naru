package cachebackend

import (
	"github.com/spikeekips/naru/common"
)

const (
	CacheItemNotFoundCode = iota + 100
)

var (
	CacheItemNotFound = common.NewError(CacheItemNotFoundCode, "item not found in cache")
)
