package sebak

import (
	"github.com/spikeekips/naru/common"
)

const (
	ProviderNotOpenedErrorCode = iota + 100
	ProviderNotClosedErrorCode
	BlockNotFoundCode
)

var (
	ProviderNotOpenedError = common.NewError(ProviderNotOpenedErrorCode, "SEBAKStorageProvider is not opened")
	ProviderNotClosedError = common.NewError(ProviderNotClosedErrorCode, "SEBAKStorageProvider is not closed")
	BlockNotFound          = common.NewError(BlockNotFoundCode, "block not found")
)
