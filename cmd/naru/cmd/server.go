package cmd

import (
	"fmt"
	"net/http/pprof"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	"github.com/spikeekips/cvc"

	restv1 "github.com/spikeekips/naru/api/rest/v1"
	cachebackend "github.com/spikeekips/naru/cache/backend"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
)

var (
	serverConfigManager *cvc.Manager
)

type ServerConfig struct {
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
	var sc *ServerConfig
	serverCmd := &cobra.Command{
		Use:  "server [<config file>]",
		Long: "naru server is the api gateway for SEBAK",
		Run: func(c *cobra.Command, args []string) {
			if len(args) > 0 {
				serverConfigManager.SetViperConfigFile(args...)
			}

			if _, err := serverConfigManager.Merge(); err != nil {
				cmdcommon.PrintError(c, err)
			}

			cs := serverConfigManager.ConfigPprint()
			cs = append(cs, "\n\tstorage-backend", sc.Storage.Backend().Type())
			log.Debug("config merged", cs...)

			SetAllLogging(sc.Log)

			log.Info("start naru server")

			if err := runServer(sc); err != nil {
				log.Error("exited with error", "error", err)
			}
		},
	}
	rootCmd.AddCommand(serverCmd)

	sc = &ServerConfig{
		SEBAK:   config.NewSEBAK(),
		Digest:  config.NewDigest(),
		System:  config.NewSystem(),
		Network: config.NewNetwork(),
		Storage: config.NewStorage(),
		Log:     config.NewLogs(),
	}
	serverConfigManager = cvc.NewManager("naru", sc, serverCmd, viper.New())
}

func runServer(sc *ServerConfig) error {
	var err error

	nodeInfo, err := getNodeInfo(sc.SEBAK.Endpoint)
	if err != nil {
		return err
	}

	log.Debug("sebak nodeinfo", "nodeinfo", nodeInfo)

	st, err := NewStorageByConfig(sc.Storage)
	if err != nil {
		return err
	}

	potion := NewPotionByStorage(st)

	if sc.Digest.Init {
		if err = st.Initialize(); err != nil {
			log.Crit("failed to remove storage", "storage", sc.Storage, "error", err)
			return err
		}
	}

	storage.Observer.On(element.EventOnAfterSaveBlock, func(v ...interface{}) {
		fmt.Println("> new block triggered", v)
	})

	storage.Observer.On(element.EventOnAfterSaveAccount, func(v ...interface{}) {
		fmt.Println("> account saved triggered", v)
	})

	provider := sebak.NewJSONRPCStorageProvider(sc.SEBAK.JSONRpc)
	sst := sebak.NewStorage(provider)

	runner := digest.NewInitializeDigestRunner(sst, potion, nodeInfo)
	if sc.Digest.RemoteBlock > 0 {
		runner.TestLastRemoteBlock = sc.Digest.RemoteBlock
	}
	if err = runner.Run(); err != nil {
		return err
	}

	watchRunner := digest.NewWatchDigestRunner(sst, potion, nodeInfo, runner.StoredRemoteBlock().Height+1)
	watchRunner.SetInterval(sc.Digest.WatchInterval)
	go func() {
		if err := watchRunner.Run(false); err != nil {
			log.Crit("failed watchRunner", "error", err)
		}
	}()

	// start network layers
	cb := cachebackend.NewGoCache()

	restServer := restv1.NewServer(sc.Network, sst, potion, cb, nodeInfo)
	if sc.System.Profile {
		restServer.AddHandler("/debug/pprof/", pprof.Index)
		restServer.AddHandler("/debug/pprof/cmdline", pprof.Cmdline)
		restServer.AddHandler("/debug/pprof/profile", pprof.Profile)
		restServer.AddHandler("/debug/pprof/symbol", pprof.Symbol)
		restServer.AddHandler("/debug/pprof/trace", pprof.Trace)
	}

	if err := restServer.Start(); err != nil {
		log.Crit("failed to run restServer", "error", err)
		return err
	}

	return nil
}
