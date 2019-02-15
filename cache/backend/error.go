package cachebackend

import (
	sebakerrors "boscoin.io/sebak/lib/errors"
)

const (
	CacheItemNotFoundCode = iota + 100
)

var (
	CacheItemNotFound = sebakerrors.NewError(CacheItemNotFoundCode, "item not found in cache")
)
