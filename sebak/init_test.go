package sebak

import (
	logging "github.com/inconshreveable/log15"

	"github.com/spikeekips/naru/common"
)

func init() {
	common.SetLoggingWithLogger(logging.LvlCrit, common.DefaultLogHandler, log)
}
