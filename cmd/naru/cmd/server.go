package cmd

import (
	"fmt"
	"net/http/pprof"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	sebaknode "boscoin.io/sebak/lib/node"
	sebakrunner "boscoin.io/sebak/lib/node/runner"
	sebakstorage "boscoin.io/sebak/lib/storage"

	restv1 "github.com/spikeekips/naru/api/rest/v1"
	cachebackend "github.com/spikeekips/naru/cache/backend"
	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/storage"
)

var (
	serverCmd   *cobra.Command
	flagProfile bool
)

func init() {
	serverCmd = &cobra.Command{
		Use:  "server",
		Long: "naru server is the api gateway for SEBAK",
		Run: func(c *cobra.Command, args []string) {
			log.Info("start naru server")

			parseBasicFlags(serverCmd)
			parseStorageFlags(serverCmd)
			parseSEBAKFlags(serverCmd)
			parseBindFlags(serverCmd)

			PrintParsedFlags(log, serverCmd, flagLogFormat != "json")

			if err := runServer(); err != nil {
				log.Error("exited with error", "error", err)
			}
		},
	}
	rootCmd.AddCommand(serverCmd)

	setBasicFlags(serverCmd)
	setSEBAKFlags(serverCmd)
	setBindFlags(serverCmd)
	setStorageFlags(serverCmd)
	serverCmd.Flags().BoolVar(&flagInit, "init", flagInit, "initialize")
	serverCmd.Flags().BoolVar(&flagProfile, "profile", flagProfile, "enable http profiling")
}

func runServer() error {
	var err error
	var nodeInfo sebaknode.NodeInfo
	{ // get node info
		client, err := common.NewHTTP2Client(time.Second*2, (*url.URL)(sebakEndpoint), false, nil)
		if err != nil {
			err = fmt.Errorf("failed to create network client for sebak")
			log.Crit(err.Error(), "endpoint", sebakEndpoint)
			return err
		}

		var b []byte
		if b, err = client.Get(sebakrunner.NodeInfoHandlerPattern, nil); err != nil {
			log.Crit("failed to get node info", "endpoint", sebakEndpoint, "error", err)
			return err
		}

		if nodeInfo, err = sebaknode.NewNodeInfoFromJSON(b); err != nil {
			log.Crit(
				"failed to parse node info",
				"endpoint", sebakEndpoint,
				"error", err,
				"received", string(b),
			)
			return err
		}

		log.Debug("sebak nodeinfo", "nodeinfo", nodeInfo)
	}

	if flagInit {
		if err = os.RemoveAll(storageConfig.Path); err != nil {
			cmdcommon.PrintFlagsError(digestCmd, "--storage", err)
		}
	}

	// run digest first
	var nst *sebakstorage.LevelDBBackend
	if nst, err = sebakstorage.NewStorage(storageConfig); err != nil {
		cmdcommon.PrintFlagsError(digestCmd, "--storage", err)
	}

	st = storage.NewStorage(nst)

	storage.Observer.On(storage.EventNewBlock, func(v ...interface{}) {
		fmt.Println("> new block triggered", v)
	})

	storage.Observer.On(storage.EventNewAccount, func(v ...interface{}) {
		fmt.Println("> new account triggered", v)
	})

	storage.Observer.On(storage.EventUpdateAccount, func(v ...interface{}) {
		fmt.Println("> account updated", v)
	})

	runner := digest.NewInitializeDigestRunner(st, sst, sebakInfo)
	if flagRemoteBlock > 0 {
		runner.TestLastRemoteBlock = flagRemoteBlock
	}
	if err = runner.Run(); err != nil {
		return err
	}

	watchRunner := digest.NewWatchDigestRunner(st, sst, sebakInfo, runner.StoredRemoteBlock().Height+1)
	go func() {
		if err := watchRunner.Run(false); err != nil {
			log.Crit("failed watchRunner", "error", err)
		}
	}()

	// start network layers
	cb := cachebackend.NewGoCache()

	restServer := restv1.NewServer(bindEndpoint, st, sst, cb, sebakInfo)
	if flagProfile {
		restServer.AddHandler("/debug/pprof/", pprof.Index)
		restServer.AddHandler("/debug/pprof/cmdline", pprof.Cmdline)
		restServer.AddHandler("/debug/pprof/profile", pprof.Profile)
		restServer.AddHandler("/debug/pprof/symbol", pprof.Symbol)
		restServer.AddHandler("/debug/pprof/trace", pprof.Trace)
	}

	if err := restServer.Start(); err != nil {
		log.Crit("failed to run restServer", "error", err)
	}

	return nil
}
