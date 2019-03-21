package restv1

import (
	"github.com/spikeekips/naru/common"
)

const (
	InvalidMessageCode = iota + 100
	TransactionNotFoundCode
	BadRequestParameterCode
	PageQueryLimitMaxExceedCode
)

var (
	InvalidMessage          = common.NewError(InvalidMessageCode, "invalid message")
	TransactionNotFound     = common.NewError(TransactionNotFoundCode, "transaction not found")
	BadRequestParameter     = common.NewError(BadRequestParameterCode, "found invalid request")
	PageQueryLimitMaxExceed = common.NewError(PageQueryLimitMaxExceedCode, "limit exceeded in page")
)
