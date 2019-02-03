package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	sebakerrors "boscoin.io/sebak/lib/errors"
	sebakhttputils "boscoin.io/sebak/lib/network/httputils"
	sebakresource "boscoin.io/sebak/lib/node/runner/api/resource"

	"github.com/spikeekips/naru/storage/item"
)

var (
	GetAccountsDefaultLimit int = 100
)

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["id"]

	ac, err := item.GetAccount(h.st, address)
	if err != nil {
		sebakhttputils.WriteJSONError(w, err)
		return
	}

	jw := JSONWriter{w: w}

	payload := sebakresource.NewAccount(&ac.BlockAccount)
	jw.Write(payload)
}

func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	jw := JSONWriter{w: w}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		jw.Write(err)
		return
	}

	var addresses []string
	if err := json.Unmarshal(body, &addresses); err != nil {
		jw.Write(sebakerrors.BadRequestParameter.Clone().SetData("error", err.Error()))
		return
	}
	if len(addresses) > GetAccountsDefaultLimit {
		jw.Write(sebakerrors.PageQueryLimitMaxExceed)
		return
	} else if len(addresses) < 1 {
		jw.Write(sebakerrors.BadRequestParameter)
		return
	}

	var rs []sebakresource.Resource
	for _, address := range addresses {
		ac, err := item.GetAccount(h.st, address)
		if err != nil {
			jw.Write(err)
			return
		}
		rs = append(rs, sebakresource.NewAccount(&ac.BlockAccount))
	}

	jw.Write(sebakresource.NewResourceList(rs, "", "", ""))
}
