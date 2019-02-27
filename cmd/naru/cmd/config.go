package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spikeekips/cvc"
	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/storage"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/sebak"
)

type configFileConfig struct {
	cvc.BaseGroup
	Format string `flag-help:"config format {json yml toml properties}"`
}

type configCheckConfig struct {
	cvc.BaseGroup
	Log     *config.Log
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
				printCheck(c, configCheckManager, configFileManager.Config().(*configFileConfig).Format)
				printCheck(c, configFileManager, configFileManager.Config().(*configFileConfig).Format)
				printCheck(c, serverConfigManager, configFileManager.Config().(*configFileConfig).Format)
			},
		}
		configCmd.AddCommand(configFileCmd)

		configFileManager = cvc.NewManager(
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
				config.Log.SetLogging(cvc.SetLogging)
				config.Log.SetLogging(common.SetLogging)
				config.Log.SetLogging(digest.SetLogging)
				config.Log.SetLogging(sebak.SetLogging)
				config.Log.SetLogging(storage.SetLogging)

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
