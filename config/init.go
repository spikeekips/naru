package config

import (
	logging "github.com/inconshreveable/log15"

	"github.com/spikeekips/naru/common"
)

var log logging.Logger = logging.New("module", "config")

func init() {
	common.SetLoggingWithLogger(common.DefaultLogLevel, common.DefaultLogHandler, log)
}

func SetLogging(level logging.Lvl, handler logging.Handler) {
	common.SetLoggingWithLogger(level, handler, log)
}
