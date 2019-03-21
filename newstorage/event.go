package newstorage

import (
	"github.com/spikeekips/naru/common"
)

var (
	Observer *common.Observable
)

func init() {
	Observer = common.NewObservable("storage")
}
