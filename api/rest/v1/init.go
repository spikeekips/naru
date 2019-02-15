package restv1

import (
	logging "github.com/inconshreveable/log15"

	"github.com/spikeekips/naru/common"
)

var log logging.Logger = logging.New("module", "rest")
var httpLog logging.Logger = logging.New("module", "rest")

func init() {
	common.SetLogging(common.DefaultLogLevel, common.DefaultLogHandler, log)
}

func SetLogging(level logging.Lvl, handler logging.Handler) {
	common.SetLogging(level, handler, log)
}

func SetHTTPLogging(level logging.Lvl, handler logging.Handler) {
	common.SetLogging(level, handler, httpLog)
}
