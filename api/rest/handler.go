package rest

import (
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
)

type Handler struct {
	st  *storage.Storage
	sst *sebak.Storage
}

func NewHandler(st *storage.Storage, sst *sebak.Storage) *Handler {
	return &Handler{st: st, sst: sst}
}
