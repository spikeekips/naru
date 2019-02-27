package cmd

import (
	"os"

	logging "github.com/inconshreveable/log15"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"

	restv1 "github.com/spikeekips/naru/api/rest/v1"
	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/storage"
)

var (
	log     logging.Logger = logging.New("module", "main")
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "naru",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			c.Usage()
		}
	},
}

func SetLogging(level logging.Lvl, handler logging.Handler) {
	common.SetLoggingWithLogger(level, handler, log)
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

	lvl := common.DefaultLogLevel
	if verbose {
		lvl = logging.LvlDebug
		common.SetLogging(lvl, logHandler)
		digest.SetLogging(lvl, logHandler)
		restv1.SetLogging(lvl, logHandler)
		storage.SetLogging(lvl, logHandler)
	}

	SetLogging(lvl, logHandler)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		cmdcommon.PrintFlagsError(rootCmd, "", err)
	}
}
