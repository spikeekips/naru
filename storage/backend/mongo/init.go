package mongostorage

import (
	logging "github.com/inconshreveable/log15"
)

var (
	log logging.Logger = logging.New("module", "mongostorage")
)

func Log() logging.Logger {
	return log
}
