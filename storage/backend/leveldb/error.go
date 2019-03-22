package leveldbstorage

import (
	"github.com/spikeekips/naru/common"
)

const (
	_ = iota
	LevelDBCoreErrorCode
)

var (
	LevelDBCoreError = common.NewError(LevelDBCoreErrorCode, "leveldb error")
)
