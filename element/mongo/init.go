package mongoelement

import (
	logging "github.com/inconshreveable/log15"
)

var log logging.Logger = logging.New("module", "mongoelement")

func Log() logging.Logger {
	return log
}
