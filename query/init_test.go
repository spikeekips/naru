package query

import (
	"os"

	logging "github.com/inconshreveable/log15"
)

func init() {
	log.SetHandler(logging.LvlFilterHandler(
		//logging.LvlDebug,
		logging.LvlError,
		logging.StreamHandler(os.Stdout, logging.TerminalFormat()),
	))
}
