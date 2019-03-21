package query

import (
	"github.com/spikeekips/naru/common"
)

const (
	InvaludValueCode = iota + 200
	InvalidQueryTypeCode
	NotSupportedOperatorCode
	NotSupportedConjunctionCode
	NotSupportedValueInMongoCode
)

var (
	InvaludValue             = common.NewError(InvaludValueCode, "invalid term value")
	InvalidQueryType         = common.NewError(InvalidQueryTypeCode, "invalid query type")
	NotSupportedOperator     = common.NewError(NotSupportedOperatorCode, "not supported operator")
	NotSupportedConjunction  = common.NewError(NotSupportedConjunctionCode, "not supported conjunction")
	NotSupportedValueInMongo = common.NewError(NotSupportedValueInMongoCode, "not supported value in mongodb")
)
