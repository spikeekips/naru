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
			parseBasicFlags(digestCmd)

			log.Info("start naru")

			parseSEBAKFlags(digestCmd)
			parseStorageFlags(digestCmd)

			// print flags
			parsedFlags := []interface{}{}
			parsedFlags = append(parsedFlags, "\n\t log-level", flagLogLevel)
			parsedFlags = append(parsedFlags, "\n\t log-format", flagLogFormat)
			parsedFlags = append(parsedFlags, "\n\t log", flagLog)
			parsedFlags = append(parsedFlags, "\n\t sebak", flagSebak)
			parsedFlags = append(parsedFlags, "\n\t jsonrpc", flagJSONRPC)
			parsedFlags = append(parsedFlags, "\n\t storage", flagStorage)
			parsedFlags = append(parsedFlags, "\n\t sebak-info", sebakInfo)
			parsedFlags = append(parsedFlags, "\n\t init", flagInit)
			parsedFlags = append(parsedFlags, "\n", "")

			log.Debug("parsed flags:", parsedFlags...)

			if err := runDigest(); err != nil {
				log.Error("exited with error", "error", err)
			} else {
				log.Info("finished")
			}
		},
	}
	rootCmd.AddCommand(digestCmd)

	setBasicFlags(digestCmd)
	digestCmd.Flags().StringVar(&flagStorage, "storage", flagStorage, "storage uri")
	setSEBAKFlags(digestCmd)
	digestCmd.Flags().BoolVar(&flagInit, "init", flagInit, "initialize")
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

	runner := digest.NewInitializeDigestRunner(st, sst)
	if flagRemoteBlock > 0 {
		runner.TestLastRemoteBlock = flagRemoteBlock
	}
	if err = runner.Run(); err != nil {
		return err
	}

	watchRunner := digest.NewWatchDigestRunner(st, sst, runner.StoredRemoteBlock().Height+1)
	if err = watchRunner.Run(); err != nil {
		return err
	}

	return nil
}
