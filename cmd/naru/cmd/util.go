package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebaknode "boscoin.io/sebak/lib/node"
	sebakrunner "boscoin.io/sebak/lib/node/runner"

	restv1 "github.com/spikeekips/naru/api/rest/v1"
	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/element"
	leveldbitem "github.com/spikeekips/naru/element/leveldb"
	mongoitem "github.com/spikeekips/naru/element/mongo"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
	storagebackend "github.com/spikeekips/naru/storage/backend"
	leveldbstorage "github.com/spikeekips/naru/storage/backend/leveldb"
	mongostorage "github.com/spikeekips/naru/storage/backend/mongo"
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

func SetAllLogging(c *config.Logs) {
	c.Global.SetLogger(Log())

	c.Package.Config.SetLogger(config.Log())
	c.Package.Common.SetLogger(common.Log())
	c.Package.Digest.SetLogger(digest.Log())
	c.Package.Restv1.SetLogger(restv1.Log())
	c.Package.SEBAK.SetLogger(sebak.Log())
	c.Package.Storage.SetLogger(storage.Log())
	c.Package.StorageBackend.SetLogger(storagebackend.Log())
	c.Package.Query.SetLogger(storage.Log())
}

func NewStorageByConfig(c *config.Storage) (storage.Storage, error) {
	var st storage.Storage
	var err error

	switch c.Backend().Type() {
	case "mongo":
		if st, err = mongostorage.NewStorage(c.Mongo); err != nil {
			log.Crit("failed to load storage", "config", c, "error", err)
			return nil, err
		}
		mongoitem.EventSync()
	case "leveldb":
		if st, err = leveldbstorage.NewStorage(c.LevelDB); err != nil {
			log.Crit("failed to load storage", "config", c, "error", err)
			return nil, err
		}
		leveldbitem.EventSync()
	}

	return st, nil
}

func NewPotionByStorage(st storage.Storage) element.Potion {
	switch st.(type) {
	case *mongostorage.Storage:
		return mongoitem.NewPotion(st.(*mongostorage.Storage))
	case *leveldbstorage.Storage:
		return leveldbitem.NewPotion(st.(*leveldbstorage.Storage))
	default:
		panic(errors.New("invalid storage type found"))
	}

	return nil
}
