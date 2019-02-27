package cmd

import (
	"fmt"
	"net/url"
	"time"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebaknode "boscoin.io/sebak/lib/node"
	sebakrunner "boscoin.io/sebak/lib/node/runner"

	"github.com/spikeekips/naru/common"
)

func getNodeInfo(endpoint *sebakcommon.Endpoint) (sebaknode.NodeInfo, error) {
	client, err := common.NewHTTP2Client(time.Second*2, (*url.URL)(endpoint), false, nil)
	if err != nil {
		err := fmt.Errorf("failed to create network client for sebak")
		log.Crit(err.Error(), "endpoint", endpoint)
		return sebaknode.NodeInfo{}, err
	}

	b, err := client.Get(sebakrunner.NodeInfoHandlerPattern, nil)
	if err != nil {
		log.Crit("failed to get node info", "endpoint", endpoint, "error", err)
		return sebaknode.NodeInfo{}, err
	}

	nodeInfo, err := sebaknode.NewNodeInfoFromJSON(b)
	if err != nil {
		log.Crit(
			"failed to parse node info",
			"endpoint", endpoint,
			"error", err,
			"received", string(b),
		)
		return sebaknode.NodeInfo{}, err
	}

	return nodeInfo, nil
}
