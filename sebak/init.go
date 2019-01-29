package sebak

import (
	logging "github.com/inconshreveable/log15"

	"github.com/spikeekips/naru/common"
)

var log logging.Logger = logging.New("module", "sebak")

func init() {
	common.SetLogging(common.DefaultLogLevel, common.DefaultLogHandler, log)
}

func SetLogging(level logging.Lvl, handler logging.Handler) {
	common.SetLogging(level, handler, log)
}
