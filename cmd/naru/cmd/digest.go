package cmd

import (
	"os"

	"github.com/spf13/cobra"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/storage"
)

var (
	digestCmd       *cobra.Command
	flagRemoteBlock uint64 = 0
)

func init() {
	digestCmd = &cobra.Command{
		Use:  "digest",
		Long: "digest blocks from SEBAK",
		Run: func(c *cobra.Command, args []string) {
			log.Info("start naru digest")

			parseBasicFlags(digestCmd)
			parseSEBAKFlags(digestCmd)
			parseStorageFlags(digestCmd)

			PrintParsedFlags(log, digestCmd, flagLogFormat != "json")

			if err := runDigest(); err != nil {
				log.Error("exited with error", "error", err)
			} else {
				log.Info("finished")
			}
		},
	}
	rootCmd.AddCommand(digestCmd)

	setBasicFlags(digestCmd)
	setStorageFlags(digestCmd)
	setSEBAKFlags(digestCmd)

	digestCmd.Flags().BoolVar(&flagInit, "init", flagInit, "initialize")
	digestCmd.Flags().BoolVar(&flagWatch, "watch", flagWatch, "watch new block from sebak")
	digestCmd.Flags().Uint64Var(&flagRemoteBlock, "test-remote-block", flagRemoteBlock, "set last remote block for test")
}

func runDigest() error {
	var err error
	if flagInit {
		if err = os.RemoveAll(storageConfig.Path); err != nil {
			cmdcommon.PrintFlagsError(digestCmd, "--storage", err)
		}
	}

	var nst *sebakstorage.LevelDBBackend
	if nst, err = sebakstorage.NewStorage(storageConfig); err != nil {
		cmdcommon.PrintFlagsError(digestCmd, "--storage", err)
	}

	st = storage.NewStorage(nst)

	runner := digest.NewInitializeDigestRunner(st, sst, sebakInfo)
	if flagRemoteBlock > 0 {
		runner.TestLastRemoteBlock = flagRemoteBlock
	}
	if err = runner.Run(); err != nil {
		return err
	}

	if flagWatch {
		watchRunner := digest.NewWatchDigestRunner(st, sst, sebakInfo, runner.StoredRemoteBlock().Height+1)
		if err = watchRunner.Run(true); err != nil {
			return err
		}
	}

	return nil
}
