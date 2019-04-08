package graphqlapiv1

import (
	logging "github.com/inconshreveable/log15"
)

var log logging.Logger = logging.New("module", "graphqlapiv1")

func Log() logging.Logger {
	return log
}
