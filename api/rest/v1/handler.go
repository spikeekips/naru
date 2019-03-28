package restv1

import (
	sebaknode "boscoin.io/sebak/lib/node"

	"github.com/spikeekips/naru/cache"
	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/sebak"
)

type Handler struct {
	sst       *sebak.Storage
	potion    element.Potion
	cch       *cache.Cache
	sebakInfo sebaknode.NodeInfo
}

func NewHandler(sst *sebak.Storage, potion element.Potion, cch *cache.Cache, sebakInfo sebaknode.NodeInfo) *Handler {
	return &Handler{sst: sst, potion: potion, cch: cch, sebakInfo: sebakInfo}
}
