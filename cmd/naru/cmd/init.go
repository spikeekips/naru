package cmd

import (
	"os"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	logging "github.com/inconshreveable/log15"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	restv1 "github.com/spikeekips/naru/api/rest/v1"
	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
)

var (
	log     logging.Logger = logging.New("module", "main")
	verbose bool
)

func Log() logging.Logger {
	return log
}

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "naru",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			c.Usage()
		}
	},
}

func init() {
	var logFormat string = common.DefaultLogFormat
	for i, f := range os.Args[1:] {
		if !verbose && f == "--verbose" {
			verbose = true
		} else if f == "--system-log-format" && len(os.Args) > (i+1) {
			logFormat = os.Args[i+1]
		}
	}

	if verbose {
		var logHandler logging.Handler
		if logFormat == "terminal" {
			var logFormatter logging.Format
			if isatty.IsTerminal(os.Stdout.Fd()) {
				logFormatter = logging.TerminalFormat()
			} else {
				logFormatter = logging.LogfmtFormat()
			}
			logHandler = logging.StreamHandler(os.Stdout, logFormatter)
		} else {
			logHandler = common.DefaultLogHandler
		}

		Log().SetHandler(logging.LvlFilterHandler(logging.LvlDebug, logHandler))
		common.Log().SetHandler(logging.LvlFilterHandler(logging.LvlDebug, logHandler))
		digest.Log().SetHandler(logging.LvlFilterHandler(logging.LvlDebug, logHandler))
		restv1.Log().SetHandler(logging.LvlFilterHandler(logging.LvlDebug, logHandler))
		sebak.Log().SetHandler(logging.LvlFilterHandler(logging.LvlDebug, logHandler))
		storage.Log().SetHandler(logging.LvlFilterHandler(logging.LvlDebug, logHandler))
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		cmdcommon.PrintFlagsError(rootCmd, "", err)
	}
}
