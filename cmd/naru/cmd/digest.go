package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"

	"github.com/spikeekips/cvc"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/sebak"
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
				digestConfigManager.SetViperConfigFile(args...)
			}

			if _, err := digestConfigManager.Merge(); err != nil {
				cmdcommon.PrintError(c, err)
			}

			cs := serverConfigManager.ConfigPprint()
			cs = append(cs, "\n\tstorage-backend", dc.Storage.Backend().Type())
			log.Debug("config merged", cs...)

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
		SEBAK:   config.NewSEBAK(),
		Digest:  config.NewDigest(),
		System:  config.NewSystem(),
		Network: config.NewNetwork(),
		Storage: config.NewStorage(),
		Log:     config.NewLogs(),
	}
	digestConfigManager = cvc.NewManager("naru", dc, digestCmd, viper.New())
}

func runDigest(dc *digestConfig) error {
	nodeInfo, err := getNodeInfo(dc.SEBAK.Endpoint)
	if err != nil {
		return err
	}
	log.Debug("sebak nodeinfo", "nodeinfo", nodeInfo)

	st, err := NewStorageByConfig(dc.Storage)
	if err != nil {
		return err
	}

	if dc.Digest.Init {
		if err = st.Initialize(); err != nil {
			log.Crit("failed to remove storage", "storage", dc.Storage, "error", err)
			return err
		}
	}

	potion := NewPotionByStorage(st)

	provider := sebak.NewJSONRPCStorageProvider(dc.SEBAK.JSONRpc)
	sst := sebak.NewStorage(provider)

	runner := digest.NewInitializeDigestRunner(sst, potion, nodeInfo)
	if dc.Digest.RemoteBlock > 0 {
		runner.TestLastRemoteBlock = dc.Digest.RemoteBlock
	}
	if err = runner.Run(); err != nil {
		return err
	}

	if dc.Digest.Watch {
		watchRunner := digest.NewWatchDigestRunner(sst, potion, nodeInfo, runner.StoredRemoteBlock().Height+1)
		watchRunner.SetInterval(dc.Digest.WatchInterval)
		if err = watchRunner.Run(true); err != nil {
			return err
		}
	}

	return nil
}
