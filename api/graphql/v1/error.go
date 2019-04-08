package graphqlapiv1

import (
	"github.com/spikeekips/naru/common"
)

const (
	_ = iota
	PotionIsMissingCode
	InValidPublicAddressCode
	SourceNotFoundCode
	InValidArgumentCode
)

var (
	PotionIsMissing      = common.NewError(PotionIsMissingCode, "potion is missing")
	InValidPublicAddress = common.NewError(InValidPublicAddressCode, "invalid public address")
	SourceNotFound       = common.NewError(SourceNotFoundCode, "source not found in graphql params")
	InValidArgument      = common.NewError(InValidArgumentCode, "invalid argument")
)
