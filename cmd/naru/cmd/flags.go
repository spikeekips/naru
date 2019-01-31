package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	logging "github.com/inconshreveable/log15"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cmdcommon "boscoin.io/sebak/cmd/sebak/common"
	sebakcommon "boscoin.io/sebak/lib/common"
	sebaknode "boscoin.io/sebak/lib/node"
	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/naru/api/rest"
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
	sebakInfo       sebaknode.NodeInfo
	storageConfig   *sebakstorage.Config
)

var (
	flagLog         string = sebakcommon.GetENVValue("NARU_LOG", "")
	flagLogFormat   string = sebakcommon.GetENVValue("NARU_LOG_FORMAT", "json")
	flagLogLevel    string = sebakcommon.GetENVValue("NARU_LOG_LEVEL", common.DefaultLogLevel.String())
	flagHTTPLog     string = sebakcommon.GetENVValue("NARU_HTTP_LOG", "")
	flagBind        string = sebakcommon.GetENVValue("NARU_BIND", "http://0.0.0.0:23456")
	flagSebak       string = sebakcommon.GetENVValue("NARU_SEBAK", "http://localhost:12345")
	flagJSONRPC     string = sebakcommon.GetENVValue("NARU_JSONRPC", sebakcommon.DefaultJSONRPCBindURL)
	flagTLSCertFile string = sebakcommon.GetENVValue("NARU_TLS_CERT", "naru.crt")
	flagTLSKeyFile  string = sebakcommon.GetENVValue("NARU_TLS_KEY", "naru.key")
	flagStorage     string = sebakcommon.GetENVValue("NARU_STORAGE", "")
	flagInit        bool
	flagWatch       bool
)

func PrintParsedFlags(logger logging.Logger, c *cobra.Command, readable bool) {
	formatName := func(name string) string {
		if !readable {
			return name
		}
		return fmt.Sprintf("\n\t %s", name)
	}

	var pf []interface{}
	c.Flags().VisitAll(func(f *pflag.Flag) {
		pf = append(
			pf,
			formatName(f.Name),
			f.Value,
		)
	})
	pf = append(pf, "\n", "")

	logger.Debug("parsed flags:", pf...)
}

func setBasicFlags(c *cobra.Command) {
	c.Flags().StringVar(&flagLogLevel, "log-level", flagLogLevel, "log level, {crit, error, warn, info, debug}")
	c.Flags().StringVar(&flagLogFormat, "log-format", flagLogFormat, "log format, {terminal, json}")
	c.Flags().StringVar(&flagLog, "log", flagLog, "set log file")
}

func setSEBAKFlags(c *cobra.Command) {
	c.Flags().StringVar(&flagSebak, "sebak", flagSebak, "sebak endpoint")
	c.Flags().StringVar(&flagJSONRPC, "jsonrpc", flagJSONRPC, "sebak jsonrpc endpoint")
}

func setBindFlags(c *cobra.Command) {
	c.Flags().StringVar(&flagBind, "bind", flagBind, "bind to listen on")
	c.Flags().StringVar(&flagHTTPLog, "http-log", flagHTTPLog, "set log file for HTTP request")
	c.Flags().StringVar(&flagTLSCertFile, "tls-cert", flagTLSCertFile, "tls certificate file")
	c.Flags().StringVar(&flagTLSKeyFile, "tls-key", flagTLSKeyFile, "tls key file")
}

func setStorageFlags(c *cobra.Command) {
	c.Flags().StringVar(&flagStorage, "storage", flagStorage, "storage uri")
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
		digest.SetLogging(logLevel, logHandler)
		rest.SetLogging(logLevel, logHandler)
		sebak.SetLogging(logLevel, logHandler)
		storage.SetLogging(logLevel, logHandler)

		// --http-log
		if len(flagHTTPLog) < 1 {
			rest.SetHTTPLogging(logLevel, logHandler)
		} else {
			// in `http-log`, http log will be json format
			httpLogHandler, err := logging.FileHandler(flagHTTPLog, sebakcommon.JsonFormatEx(false, true))
			if err != nil {
				cmdcommon.PrintFlagsError(c, "--http-log", err)
			}
			rest.SetHTTPLogging(logging.LvlDebug, httpLogHandler) // httpLog only use `Debug`
		}
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

func parseBindFlags(c *cobra.Command) {
	log.Debug("trying to check --bind")
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
		values := bindEndpoint.Query()
		values.Set("tls-cert", flagTLSCertFile)
		values.Set("tls-key", flagTLSKeyFile)
		bindEndpoint.RawQuery = values.Encode()
	}
}
