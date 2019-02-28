package sebak

import (
	logging "github.com/inconshreveable/log15"
)

var log logging.Logger = logging.New("module", "sebak")

func Log() logging.Logger {
	return log
}
