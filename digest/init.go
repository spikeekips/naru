package digest

import (
	logging "github.com/inconshreveable/log15"
)

var log logging.Logger = logging.New("module", "digest")

func Log() logging.Logger {
	return log
}
