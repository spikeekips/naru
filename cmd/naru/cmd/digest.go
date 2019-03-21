package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"

	"github.com/spikeekips/cvc"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/digest"
	mongostorage "github.com/spikeekips/naru/newstorage/backend/mongo"
	mongoitem "github.com/spikeekips/naru/newstorage/item/mongo"
	"github.com/spikeekips/naru/sebak"
)

var (
	digestConfigManager *cvc.Manager
)

type digestConfig struct {
	cvc.BaseGroup
	SEBAK      *config.SEBAK
	Digest     *config.Digest
	System     *config.System
	Network    *config.Network
	Storage    *config.Storage
	NewStorage *config.NewStorage
	Log        *config.Logs

	Verbose bool `flag-help:"verbose"`
}

func init() {
	var dc *digestConfig
	digestCmd := &cobra.Command{
		Use:  "digest",
		Long: "digest blocks from SEBAK",
		Run: func(c *cobra.Command, args []string) {
			if len(args) > 0 {
				digestConfigManager.SetViperConfigFile(args...)
			}

			if _, err := digestConfigManager.Merge(); err != nil {
				cmdcommon.PrintError(c, err)
			}

			log.Debug("config merged", digestConfigManager.ConfigPprint()...)

			SetAllLogging(dc.Log)

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
		SEBAK:      config.NewSEBAK(),
		Digest:     config.NewDigest(),
		System:     config.NewSystem(),
		Network:    config.NewNetwork(),
		Storage:    config.NewStorage0(),
		NewStorage: config.NewNewStorage(),
		Log:        config.NewLogs(),
	}
	digestConfigManager = cvc.NewManager("naru", dc, digestCmd, viper.New())
}

func runDigest(dc *digestConfig) error {
	nodeInfo, err := getNodeInfo(dc.SEBAK.Endpoint)
	if err != nil {
		return err
	}
	log.Debug("sebak nodeinfo", "nodeinfo", nodeInfo)

	if dc.Digest.Init {
		if err = os.RemoveAll(dc.Storage.LevelDB.Path); err != nil {
			log.Crit("failed to remove storage", "directory", dc.Storage.LevelDB.Path, "error", err)
			return err
		}
	}

	//st, err := leveldbstorage.NewStorage(dc.NewStorage.LevelDB)
	st, err := mongostorage.NewStorage(dc.NewStorage.Mongo)
	if err != nil {
		log.Crit("failed to load storage", "config", dc.NewStorage, "error", err)
		return err
	}
	defer st.Close()

	mongoitem.EventSync()

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
		watchRunner.SetInterval(dc.Digest.WatchInterval)
		if err = watchRunner.Run(true); err != nil {
			return err
		}
	}

	return nil
}
