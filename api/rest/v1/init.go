package restv1

import (
	logging "github.com/inconshreveable/log15"
)

var log logging.Logger = logging.New("module", "restv1")

func Log() logging.Logger {
	return log
}
