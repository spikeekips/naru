package newstorage

import (
	"github.com/spikeekips/naru/common"
)

const (
	_                    = iota
	NotFilterBackendCode = iota
	AlreadyExistsCode
	NotFoundCode
	NotBatchStorageCode
	DecodeValueFailedCode
)

var (
	NotFilterBackend  = common.NewError(NotFilterBackendCode, "not FilterBackend")
	AlreadyExists     = common.NewError(AlreadyExistsCode, "already exists")
	NotFound          = common.NewError(NotFoundCode, "not found")
	NotBatchStorage   = common.NewError(NotBatchStorageCode, "not BatchStorage")
	DecodeValueFailed = common.NewError(DecodeValueFailedCode, "failed to decode value")
)
