package cache

import (
	logging "github.com/inconshreveable/log15"
)

var log logging.Logger = logging.New("module", "cache")

func Log() logging.Logger {
	return log
}
