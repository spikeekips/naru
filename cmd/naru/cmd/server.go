package cmd

import (
	"fmt"
	"net/http/pprof"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	"github.com/spikeekips/cvc"

	restv1 "github.com/spikeekips/naru/api/rest/v1"
	cachebackend "github.com/spikeekips/naru/cache/backend"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/digest"
	storage "github.com/spikeekips/naru/newstorage"
	leveldbstorage "github.com/spikeekips/naru/newstorage/backend/leveldb"
	"github.com/spikeekips/naru/newstorage/item"
	leveldbitem "github.com/spikeekips/naru/newstorage/item/leveldb"
	"github.com/spikeekips/naru/sebak"
)

var (
	serverConfigManager *cvc.Manager
)

type ServerConfig struct {
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

			log.Debug("config merged", serverConfigManager.ConfigPprint()...)

			SetAllLogging(sc.Log)

			log.Info("start naru server")

			if err := runServer(sc); err != nil {
				log.Error("exited with error", "error", err)
			}
		},
	}
	rootCmd.AddCommand(serverCmd)

	sc = &ServerConfig{
		SEBAK:      config.NewSEBAK(),
		Digest:     config.NewDigest(),
		System:     config.NewSystem(),
		Network:    config.NewNetwork(),
		Storage:    config.NewStorage0(),
		NewStorage: config.NewNewStorage(),
		Log:        config.NewLogs(),
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

	if sc.Digest.Init {
		if err = os.RemoveAll(sc.Storage.LevelDB.Path); err != nil {
			log.Crit("failed to remove storage", "directory", sc.Storage.LevelDB.Path, "error", err)
			return err
		}
	}

	st, err := leveldbstorage.NewStorage(sc.NewStorage.LevelDB)
	if err != nil {
		log.Crit("failed to load storage", "config", sc.NewStorage, "error", err)
		return err
	}
	defer st.Close()
	leveldbitem.EventSync()

	/*
		st, err := mongostorage.NewStorage(sc.NewStorage.Mongo)
		if err != nil {
			log.Crit("failed to load storage", "config", sc.NewStorage, "error", err)
			return err
		}
		defer st.Close()
		mongoitem.EventSync()
	*/

	storage.Observer.On(item.EventOnAfterSaveBlock, func(v ...interface{}) {
		fmt.Println("> new block triggered", v)
	})

	storage.Observer.On(item.EventOnAfterSaveAccount, func(v ...interface{}) {
		fmt.Println("> account saved triggered", v)
	})

	provider := sebak.NewJSONRPCStorageProvider(sc.SEBAK.JSONRpc)
	sst := sebak.NewStorage(provider)

	runner := digest.NewInitializeDigestRunner(st, sst, nodeInfo)
	if sc.Digest.RemoteBlock > 0 {
		runner.TestLastRemoteBlock = sc.Digest.RemoteBlock
	}
	if err = runner.Run(); err != nil {
		return err
	}

	watchRunner := digest.NewWatchDigestRunner(st, sst, nodeInfo, runner.StoredRemoteBlock().Height+1)
	watchRunner.SetInterval(sc.Digest.WatchInterval)
	go func() {
		if err := watchRunner.Run(false); err != nil {
			log.Crit("failed watchRunner", "error", err)
		}
	}()

	// start network layers
	cb := cachebackend.NewGoCache()

	restServer := restv1.NewServer(sc.Network, st, sst, cb, nodeInfo)
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
