package cmd

import (
	"fmt"
	"strings"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spikeekips/cvc"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
)

type configFileConfig struct {
	cvc.BaseGroup
	Format string `flag-help:"config format {json yml toml properties}"`
}

type configCheckConfig struct {
	cvc.BaseGroup
	Log     *config.LogConfig
	Verbose bool
}

func init() {
	configCmd := &cobra.Command{
		Use: "config",
		Run: func(c *cobra.Command, args []string) {
			if len(args) != 1 {
				c.Usage()
			}
		},
	}
	rootCmd.AddCommand(configCmd)

	var configFileManager *cvc.Manager
	var configCheckManager *cvc.Manager

	{
		printCheck := func(c *cobra.Command, manager *cvc.Manager, format string) {
			vc, err := manager.ViperString(format)
			if err != nil {
				cmdcommon.PrintError(c, err)
			}
			fmt.Println(vc)
		}

		configFileCmd := &cobra.Command{
			Use:   "file",
			Short: "print default viper configuration",
			Run: func(c *cobra.Command, args []string) {
				if _, err := configFileManager.Merge(); err != nil {
					log.Error("failed to merge config", "error", err)
					cmdcommon.PrintError(c, err)
				}

				format := configFileManager.Config().(*configFileConfig).Format
				printCheck(c, configCheckManager, format)
				printCheck(c, configFileManager, format)
				printCheck(c, digestConfigManager, format)
				printCheck(c, serverConfigManager, format)
			},
		}
		configCmd.AddCommand(configFileCmd)

		configFileManager = cvc.NewManager(
			"naru",
			&configFileConfig{Format: "yml"},
			configFileCmd,
			viper.New(),
		)
	}

	{
		configCheckCmd := &cobra.Command{
			Use:   "check [<config file> <stdin>]",
			Short: "check viper configuration file",
			Args:  cobra.MinimumNArgs(1),
			Run: func(c *cobra.Command, args []string) {
				config := configCheckManager.Config().(*configCheckConfig)
				//config.Log.SetLogger(cvc.SetLogging)
				config.Log.SetLogger(Log())
				config.Log.SetLogger(common.Log())
				config.Log.SetLogger(digest.Log())
				config.Log.SetLogger(sebak.Log())
				config.Log.SetLogger(storage.Log())

				for _, f := range args {
					log.Info("checking config file", "file", f)
					if err := configFileManager.SetViperConfigFile(f); err != nil {
						log.Error("failed to read config file", "file", f, "error", err)
						cmdcommon.PrintError(c, err)
					}
					if _, err := configFileManager.Merge(); err != nil {
						log.Error("failed to merge config file", "file", f, "error", err)
						cmdcommon.PrintError(c, err)
					}
					log.Info("no problem", "file", f)
				}
			},
		}
		configCmd.AddCommand(configCheckCmd)

		configCheckManager = cvc.NewManager(
			"naru",
			&configCheckConfig{Log: config.NewLog()},
			configCheckCmd,
			viper.New(),
		)
	}

	{
		printEnv := func(manager *cvc.Manager) {
			fmt.Println("$ naru", strings.Join(manager.Groups(), " "))
			for _, env := range manager.Envs() {
				fmt.Println("-", env)
			}
			fmt.Println()
		}

		configEnvCmd := &cobra.Command{
			Use:   "env",
			Short: "list env",
			Run: func(c *cobra.Command, args []string) {
				printEnv(configFileManager)
				printEnv(configCheckManager)
				printEnv(serverConfigManager)
			},
		}
		configCmd.AddCommand(configEnvCmd)
	}
}
