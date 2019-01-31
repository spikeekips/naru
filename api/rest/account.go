package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	sebakhttputils "boscoin.io/sebak/lib/network/httputils"
	sebakresource "boscoin.io/sebak/lib/node/runner/api/resource"

	"github.com/spikeekips/naru/storage/item"
)

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["id"]

	ac, err := item.GetAccount(h.st, address)
	if err != nil {
		sebakhttputils.WriteJSONError(w, err)
		return
	}

	jw := JSONWriter{w: w, status: 200}

	payload := sebakresource.NewAccount(&ac.BlockAccount)
	jw.Write(payload)
}
