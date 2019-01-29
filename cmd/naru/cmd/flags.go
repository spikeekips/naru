package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	logging "github.com/inconshreveable/log15"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	sebakcommon "boscoin.io/sebak/lib/common"
	sebaknetwork "boscoin.io/sebak/lib/network"
	sebaknode "boscoin.io/sebak/lib/node"
	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
)

var (
	bindEndpoint    *sebakcommon.Endpoint
	sebakEndpoint   *sebakcommon.Endpoint
	jsonrpcEndpoint *sebakcommon.Endpoint
	st              *storage.Storage
	sst             *sebak.Storage
	sn              sebaknetwork.Network
	sebakInfo       sebaknode.NodeInfo
	storageConfig   *sebakstorage.Config
)

var (
	flagLog         string = sebakcommon.GetENVValue("NARU_LOG", "")
	flagLogFormat   string = sebakcommon.GetENVValue("NARU_LOG_FORMAT", "json")
	flagLogLevel    string = sebakcommon.GetENVValue("NARU_LOG_LEVEL", common.DefaultLogLevel.String())
	flagHTTPLog     string = sebakcommon.GetENVValue("NARU_HTTP_LOG", "")
	flagBind        string = sebakcommon.GetENVValue("NARU_BIND", "https://0.0.0.0:23456")
	flagSebak       string = sebakcommon.GetENVValue("NARU_SEBAK", "http://localhost:12345")
	flagJSONRPC     string = sebakcommon.GetENVValue("NARU_JSONRPC", sebakcommon.DefaultJSONRPCBindURL)
	flagTLSCertFile string = sebakcommon.GetENVValue("NARU_TLS_CERT", "naru.crt")
	flagTLSKeyFile  string = sebakcommon.GetENVValue("NARU_TLS_KEY", "naru.key")
	flagStorage     string = sebakcommon.GetENVValue("NARU_STORAGE", "")
	flagInit        bool
)

func setBasicFlags(c *cobra.Command) {
	c.Flags().StringVar(&flagLogLevel, "log-level", flagLogLevel, "log level, {crit, error, warn, info, debug}")
	c.Flags().StringVar(&flagLogFormat, "log-format", flagLogFormat, "log format, {terminal, json}")
	c.Flags().StringVar(&flagLog, "log", flagLog, "set log file")
}

func setSEBAKFlags(c *cobra.Command) {
	c.Flags().StringVar(&flagSebak, "sebak", flagSebak, "sebak endpoint")
	c.Flags().StringVar(&flagJSONRPC, "jsonrpc", flagJSONRPC, "sebak jsonrpc endpoint")
}

func parseBasicFlags(c *cobra.Command) {
	{ // --log*
		var err error

		var logLevel logging.Lvl
		if logLevel, err = logging.LvlFromString(flagLogLevel); err != nil {
			cmdcommon.PrintFlagsError(c, "--log-level", err)
		}

		var logFormatter logging.Format
		switch flagLogFormat {
		case "terminal":
			if isatty.IsTerminal(os.Stdout.Fd()) && len(flagLog) < 1 {
				logFormatter = logging.TerminalFormat()
			} else {
				logFormatter = logging.LogfmtFormat()
			}
		case "json":
			logFormatter = sebakcommon.JsonFormatEx(false, true)
		default:
			cmdcommon.PrintFlagsError(c, "--log-format", fmt.Errorf("'%s'", flagLogFormat))
		}

		logHandler := logging.StreamHandler(os.Stdout, logFormatter)
		if len(flagLog) > 0 {
			if logHandler, err = logging.FileHandler(flagLog, logFormatter); err != nil {
				cmdcommon.PrintFlagsError(c, "--log", err)
			}
		}

		if logLevel == logging.LvlDebug { // only debug produces `caller` data
			logHandler = logging.CallerFileHandler(logHandler)
		}
		logHandler = logging.LvlFilterHandler(logLevel, logHandler)
		log.SetHandler(logHandler)

		common.SetLogging(logLevel, logHandler)
		storage.SetLogging(logLevel, logHandler)
		digest.SetLogging(logLevel, logHandler)
	}

}

func parseSEBAKFlags(c *cobra.Command) {
	{ // --sebak
		if p, err := sebakcommon.ParseEndpoint(flagSebak); err != nil {
			cmdcommon.PrintFlagsError(c, "--sebak", err)
		} else {
			sebakEndpoint = p
		}

		{
			log.Debug("trying to check --sebak", "sebak", flagSebak)
			networkConfig, err := sebaknetwork.NewHTTP2NetworkConfigFromEndpoint("naru", sebakEndpoint)
			if err != nil {
				cmdcommon.PrintFlagsError(c, "--sebak", err)
			}

			sn = sebaknetwork.NewHTTP2Network(networkConfig)

			client, err := common.NewHTTP2Client(
				time.Second*2,
				(*url.URL)(sebakEndpoint),
				false,
				http.Header{"Content-Type": []string{"application/json"}},
			)
			if err != nil {
				cmdcommon.PrintFlagsError(c, "--sebak", err)
			}

			if b, err := client.Get("/", nil); err != nil {
				cmdcommon.PrintFlagsError(c, "--sebak", err)
			} else {
				if sebakInfo, err = sebaknode.NewNodeInfoFromJSON(b); err != nil {
					log.Crit(
						"failed to parse node info",
						"endpoint", sebakEndpoint,
						"error", err,
						"received", string(b),
					)
					cmdcommon.PrintFlagsError(c, "--sebak", err)
				}
			}
		}
	}

	{ // --jsonrpc
		log.Debug("trying to check --jsonrpc", "jsonrpc", flagJSONRPC)
		if p, err := sebakcommon.ParseEndpoint(flagJSONRPC); err != nil {
			cmdcommon.PrintFlagsError(c, "--jsonrpc", err)
		} else {
			jsonrpcEndpoint = p
		}

		j := sebak.NewJSONRPCStorageProvider(jsonrpcEndpoint)
		sst = sebak.NewStorage(j)
		if err := j.Open(); err != nil {
			cmdcommon.PrintFlagsError(c, "--jsonrpc", err)
		} else if err := j.Close(); err != nil {
			cmdcommon.PrintFlagsError(c, "--jsonrpc", err)
		}
	}
}

func parseStorageFlags(c *cobra.Command) {
	if len(flagStorage) < 1 {
		flagStorage = cmdcommon.GetDefaultStoragePath(c)
	}

	var err error
	if storageConfig, err = sebakstorage.NewConfigFromString(flagStorage); err != nil {
		cmdcommon.PrintFlagsError(c, "--storage", err)
	}
}
