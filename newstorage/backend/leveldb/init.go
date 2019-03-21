package leveldbstorage

import (
	logging "github.com/inconshreveable/log15"
)

var (
	log logging.Logger = logging.New("module", "leveldbstorage")
)

func Log() logging.Logger {
	return log
}
