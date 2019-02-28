package common

import (
	"os"
	"time"

	logging "github.com/inconshreveable/log15"

	sebakcommon "boscoin.io/sebak/lib/common"
)

var (
	DefaultLogFormat           string          = "json"
	DefaultLogHandler          logging.Handler = logging.StreamHandler(os.Stdout, logging.JsonFormat())
	DefaultLogLevel            logging.Lvl     = logging.LvlError
	DefaultSEBAKEndpointString string          = "http://localhost:12345"
	DefaultSEBAKEndpoint       *sebakcommon.Endpoint
	DefaultSEBAKJSONRpcString  string = "http://localhost:54321/jsonrpc"
	DefaultSEBAKJSONRpc        *sebakcommon.Endpoint
	DefaultBindString          string = "http://0.0.0.0:23456"
	DefaultBind                *sebakcommon.Endpoint
	DefaultTLSCert             string        = "naru.crt"
	DefaultTLSKey              string        = "naru.key"
	DefaultStoragePath         string        = "db"
	DefaultDigestWatchInterval time.Duration = time.Millisecond * 1000
)

func init() {
	DefaultSEBAKEndpoint, _ = sebakcommon.ParseEndpoint(DefaultSEBAKEndpointString)
	DefaultSEBAKJSONRpc, _ = sebakcommon.ParseEndpoint(DefaultSEBAKJSONRpcString)
	DefaultBind, _ = sebakcommon.ParseEndpoint(DefaultBindString)
}
