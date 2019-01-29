package common

import (
	"os"

	logging "github.com/inconshreveable/log15"

	sebakcommon "boscoin.io/sebak/lib/common"
)

var (
	log logging.Logger = logging.New("module", "common")

	DefaultLogLevel   logging.Lvl     = logging.LvlInfo
	DefaultLogHandler logging.Handler = logging.StreamHandler(os.Stdout, logging.JsonFormat())
)

func init() {
	SetLogging(sebakcommon.DefaultLogLevel, sebakcommon.DefaultLogHandler, log)
}

func SetLogging(level logging.Lvl, handler logging.Handler, logger ...logging.Logger) {
	if len(logger) < 1 {
		logger = []logging.Logger{log}
	}

	logger[0].SetHandler(logging.LvlFilterHandler(level, handler))
}
