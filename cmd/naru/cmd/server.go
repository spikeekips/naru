package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	sebakcommon "boscoin.io/sebak/lib/common"
	sebaknetwork "boscoin.io/sebak/lib/network"
	sebaknode "boscoin.io/sebak/lib/node"
)

var (
	serverCmd *cobra.Command
)

func init() {
	serverCmd = &cobra.Command{
		Use:  "server",
		Long: "naru server is the api gateway for SEBAK",
		Run: func(c *cobra.Command, args []string) {
			parseBasicFlags(serverCmd)
			parseSEBAKFlags(serverCmd)
			parseBindFlags(serverCmd)

			if err := run(); err != nil {
				log.Error("exited with error", "error", err)
			}
		},
	}
	rootCmd.AddCommand(serverCmd)

	setBasicFlags(serverCmd)
	setSEBAKFlags(serverCmd)

	serverCmd.Flags().StringVar(&flagBind, "bind", flagBind, "bind to listen on")
	serverCmd.Flags().StringVar(&flagHTTPLog, "http-log", flagHTTPLog, "set log file for HTTP request")
	serverCmd.Flags().StringVar(&flagTLSCertFile, "tls-cert", flagTLSCertFile, "tls certificate file")
	serverCmd.Flags().StringVar(&flagTLSKeyFile, "tls-key", flagTLSKeyFile, "tls key file")
}

func parseBindFlags(c *cobra.Command) {
	// --bind
	if p, err := sebakcommon.ParseEndpoint(flagBind); err != nil {
		cmdcommon.PrintFlagsError(c, "--bind", err)
	} else {
		bindEndpoint = p
	}

	if strings.ToLower(bindEndpoint.Scheme) == "https" {
		if _, err := os.Stat(flagTLSCertFile); os.IsNotExist(err) {
			cmdcommon.PrintFlagsError(c, "--tls-cert", err)
		}
		if _, err := os.Stat(flagTLSKeyFile); os.IsNotExist(err) {
			cmdcommon.PrintFlagsError(c, "--tls-key", err)
		}
	}
}

func run() error {
	var nt sebaknetwork.Network
	{
		networkConfig, err := sebaknetwork.NewHTTP2NetworkConfigFromEndpoint("naru", bindEndpoint)
		if err != nil {
			log.Crit("failed to create network", "error", err)
			return err
		}

		nt = sebaknetwork.NewHTTP2Network(networkConfig)
	}

	var nodeInfo sebaknode.NodeInfo
	{ // node info
		var err error

		client := nt.GetClient(sebakEndpoint)
		if client == nil {
			err = fmt.Errorf("failed to create network client for sebak")
			log.Crit(err.Error(), "endpoint", sebakEndpoint)
			return err
		}
		var b []byte
		if b, err = client.GetNodeInfo(); err != nil {
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

		fmt.Println(">>>>", nodeInfo)
	}

	{ // check jsonrpc
	}

	// validates the existing data

	// bind network

	return nil
}
