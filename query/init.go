package query

import (
	logging "github.com/inconshreveable/log15"
)

var (
	log logging.Logger = logging.New("module", "query")
)

func Log() logging.Logger {
	return log
}
