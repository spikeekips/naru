package digest

import (
	logging "github.com/inconshreveable/log15"

	"github.com/spikeekips/naru/common"
)

func init() {
	common.SetLogging(logging.LvlCrit, common.DefaultLogHandler, log)
}
