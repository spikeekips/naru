package restv1

import (
	sebaknode "boscoin.io/sebak/lib/node"

	"github.com/spikeekips/naru/cache"
	storage "github.com/spikeekips/naru/newstorage"
	"github.com/spikeekips/naru/sebak"
)

type Handler struct {
	st        storage.Storage
	sst       *sebak.Storage
	cch       *cache.Cache
	sebakInfo sebaknode.NodeInfo
}

func NewHandler(st storage.Storage, sst *sebak.Storage, cch *cache.Cache, sebakInfo sebaknode.NodeInfo) *Handler {
	return &Handler{st: st, sst: sst, cch: cch, sebakInfo: sebakInfo}
}
