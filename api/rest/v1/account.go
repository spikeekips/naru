package restv1

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	sebakresource "boscoin.io/sebak/lib/node/runner/api/resource"

	"github.com/spikeekips/naru/api/rest"
	"github.com/spikeekips/naru/newstorage/item"
)

var (
	GetAccountsDefaultLimit int = 100
)

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["id"]

	jw := rest.NewJSONWriter(w)

	ac, err := item.GetAccount(h.st, address)
	if err != nil {
		jw.WriteObject(err)
		return
	}

	payload := sebakresource.NewAccount(&ac.BlockAccount)
	jw.WriteObject(payload)
}

func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	jw := rest.NewJSONWriter(w)

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
		ac, err := item.GetAccount(h.st, address)
		if err != nil {
			jw.WriteObject(err)
			return
		}
		rs = append(rs, sebakresource.NewAccount(&ac.BlockAccount))
	}

	jw.WriteObject(sebakresource.NewResourceList(rs, "", "", ""))
}
