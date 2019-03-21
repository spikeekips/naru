package mongostorage

import (
	"github.com/spikeekips/naru/common"
)

const (
	_ = iota
	InvalidDocumentKeyCode
	InvalidDocumentValueCode
)

var (
	InvalidDocumentKey   = common.NewError(InvalidDocumentKeyCode, "document key is invalid")
	InvalidDocumentValue = common.NewError(InvalidDocumentValueCode, "document value is invalid")
)
