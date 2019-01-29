package cmd

import (
	"os"

	logging "github.com/inconshreveable/log15"
	"github.com/spf13/cobra"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
)

var (
	log logging.Logger = logging.New("module", "main")
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		cmdcommon.PrintFlagsError(rootCmd, "", err)
	}
}
