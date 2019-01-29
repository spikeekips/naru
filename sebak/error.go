package sebak

import (
	sebakerrors "boscoin.io/sebak/lib/errors"
)

const (
	ProviderNotOpenedErrorCode = iota + 100
	ProviderNotClosedErrorCode
)

var (
	ProviderNotOpenedError = sebakerrors.NewError(ProviderNotOpenedErrorCode, "SEBAKStorageProvider is not opened")
	ProviderNotClosedError = sebakerrors.NewError(ProviderNotClosedErrorCode, "SEBAKStorageProvider is not closed")
)
