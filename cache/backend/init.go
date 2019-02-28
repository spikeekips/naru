package cachebackend

import (
	logging "github.com/inconshreveable/log15"
)

var log logging.Logger = logging.New("module", "cachebackend")

func Log() logging.Logger {
	return log
}
