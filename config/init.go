package config

import (
	logging "github.com/inconshreveable/log15"
)

var log logging.Logger = logging.New("module", "config")

func Log() logging.Logger {
	return log
}
