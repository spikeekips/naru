package rest

import (
	sebaknode "boscoin.io/sebak/lib/node"

	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
)

type Handler struct {
	st        *storage.Storage
	sst       *sebak.Storage
	sebakInfo sebaknode.NodeInfo
}

func NewHandler(st *storage.Storage, sst *sebak.Storage, sebakInfo sebaknode.NodeInfo) *Handler {
	return &Handler{st: st, sst: sst, sebakInfo: sebakInfo}
}
