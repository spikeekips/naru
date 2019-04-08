package restv1

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	sebakresource "boscoin.io/sebak/lib/node/runner/api/resource"
	"github.com/gorilla/mux"

	"github.com/spikeekips/naru/api/rest"
)

var (
	GetAccountsDefaultLimit int = 100
)

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["id"]

	jw := rest.NewJSONWriter(w, r)

	ac, err := h.potion.Account(address)
	if err != nil {
		jw.WriteObject(err)
		return
	}

	payload := sebakresource.NewAccount(ac.BlockAccount())
	jw.WriteObject(payload)
}

func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	jw := rest.NewJSONWriter(w, r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		jw.WriteObject(err)
		return
	}

	var addresses []string
	if err := json.Unmarshal(body, &addresses); err != nil {
		jw.WriteObject(BadRequestParameter.New().SetData("error", err.Error()))
		return
	}
	if len(addresses) > GetAccountsDefaultLimit {
		jw.WriteObject(PageQueryLimitMaxExceed)
		return
	} else if len(addresses) < 1 {
		jw.WriteObject(BadRequestParameter)
		return
	}

	var rs []sebakresource.Resource
	for _, address := range addresses {
		ac, err := h.potion.Account(address)
		if err != nil {
			jw.WriteObject(err)
			return
		}
		rs = append(rs, sebakresource.NewAccount(ac.BlockAccount()))
	}

	jw.WriteObject(sebakresource.NewResourceList(rs, "", "", ""))
}
