package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/cvc"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
)

var (
	digestConfigManager *cvc.Manager
)

type digestConfig struct {
	cvc.BaseGroup
	SEBAK   *config.SEBAK
	Digest  *config.Digest
	System  *config.System
	Network *config.Network
	Storage *config.Storage
	Log     *config.Logs

	Verbose bool `flag-help:"verbose"`
}

func init() {
	var dc *digestConfig
	digestCmd := &cobra.Command{
		Use:  "digest",
		Long: "digest blocks from SEBAK",
		Run: func(c *cobra.Command, args []string) {
			if len(args) > 0 {
				serverConfigManager.SetViperConfigFile(args...)
			}

			if _, err := digestConfigManager.Merge(); err != nil {
				cmdcommon.PrintError(c, err)
			}

			if err := dc.Log.SetAllLogging(log); err != nil {
				cmdcommon.PrintError(c, err)
			}

			log.Info("start naru digest")

			if err := runDigest(dc); err != nil {
				log.Error("exited with error", "error", err)
			} else {
				log.Info("finished")
			}
		},
	}
	rootCmd.AddCommand(digestCmd)

	dc = &digestConfig{
		SEBAK:   config.NewSEBAK(),
		Digest:  &config.Digest{},
		System:  config.NewSystem(),
		Network: config.NewNetwork(),
		Storage: config.NewStorage(),
		Log:     config.NewLogs(),
	}
	digestConfigManager = cvc.NewManager(dc, digestCmd, viper.New())

}

func runDigest(dc *digestConfig) error {
	nodeInfo, err := getNodeInfo(dc.SEBAK.Endpoint)
	if err != nil {
		return err
	}
	log.Debug("sebak nodeinfo", "nodeinfo", nodeInfo)

	if dc.Digest.Init {
		if err = os.RemoveAll(dc.Storage.Path); err != nil {
			log.Crit("failed to remove storage", "directory", dc.Storage.Path, "error", err)
			return err
		}
	}

	var nst *sebakstorage.LevelDBBackend
	if nst, err = sebakstorage.NewStorage(dc.Storage.StorageConfig()); err != nil {
		log.Crit("failed to load leveldb storage", "directory", dc.Storage.Path, "error", err)
		return err
	}

	st := storage.NewStorage(nst)
	provider := sebak.NewJSONRPCStorageProvider(dc.SEBAK.JSONRpc)
	sst := sebak.NewStorage(provider)

	runner := digest.NewInitializeDigestRunner(st, sst, nodeInfo)
	if dc.Digest.RemoteBlock > 0 {
		runner.TestLastRemoteBlock = dc.Digest.RemoteBlock
	}
	if err = runner.Run(); err != nil {
		return err
	}

	if dc.Digest.Watch {
		watchRunner := digest.NewWatchDigestRunner(st, sst, nodeInfo, runner.StoredRemoteBlock().Height+1)
		if err = watchRunner.Run(true); err != nil {
			return err
		}
	}

	return nil
}
