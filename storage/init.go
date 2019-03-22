package storage

import (
	logging "github.com/inconshreveable/log15"
)

var log logging.Logger = logging.New("module", "storage")

func Log() logging.Logger {
	return log
}
