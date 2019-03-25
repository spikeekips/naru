package restv1

import (
	sebaknode "boscoin.io/sebak/lib/node"

	"github.com/spikeekips/naru/cache"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage/item"
)

type Handler struct {
	sst       *sebak.Storage
	getter    item.Getter
	cch       *cache.Cache
	sebakInfo sebaknode.NodeInfo
}

func NewHandler(sst *sebak.Storage, getter item.Getter, cch *cache.Cache, sebakInfo sebaknode.NodeInfo) *Handler {
	return &Handler{sst: sst, getter: getter, cch: cch, sebakInfo: sebakInfo}
}
