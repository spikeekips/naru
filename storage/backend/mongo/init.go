package mongostorage

import (
	logging "github.com/inconshreveable/log15"

	storagebackend "github.com/spikeekips/naru/storage/backend"
)

var (
	log logging.Logger = storagebackend.Log()
)

func Log() logging.Logger {
	return log
}
