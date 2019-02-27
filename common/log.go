package common

import (
	logging "github.com/inconshreveable/log15"

	sebakcommon "boscoin.io/sebak/lib/common"
)

var (
	log logging.Logger = logging.New("module", "common")
)

func init() {
	log.SetHandler(logging.LvlFilterHandler(sebakcommon.DefaultLogLevel, sebakcommon.DefaultLogHandler))
}

func SetLoggingWithLogger(level logging.Lvl, handler logging.Handler, logger logging.Logger) {
	logger.SetHandler(logging.LvlFilterHandler(level, handler))
}

func SetLogging(level logging.Lvl, handler logging.Handler) {
	log.SetHandler(logging.LvlFilterHandler(level, handler))
}
