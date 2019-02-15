package restv1

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

	sebakblock "boscoin.io/sebak/lib/block"
	sebakcommon "boscoin.io/sebak/lib/common"
	sebakerrors "boscoin.io/sebak/lib/errors"
	sebakresource "boscoin.io/sebak/lib/node/runner/api/resource"
	sebaktransaction "boscoin.io/sebak/lib/transaction"

	resourcev1 "github.com/spikeekips/naru/api/rest/v1/resource"
	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/storage/item"
)

func (h *Handler) PostTransaction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	jw := NewJSONWriter(w)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		jw.WriteObject(err)
		return
	}

	var tx sebaktransaction.Transaction
	if err := json.Unmarshal(body, &tx); err != nil {
		err = sebakerrors.InvalidMessage.Clone().SetData(
			"status", http.StatusBadRequest,
		)
		jw.WriteObject(err)
		return
	}

	conf := sebakcommon.Config{
		NetworkID: []byte(h.sebakInfo.Policy.NetworkID),
		TxsLimit:  h.sebakInfo.Policy.TransactionsLimit,
		OpsLimit:  h.sebakInfo.Policy.OperationsLimit,
	}
	if err := tx.IsWellFormed(conf); err != nil {
		jw.WriteObject(err)
		return
	}

	client, err := common.NewHTTP2Client(
		time.Second*2,
		(*url.URL)(h.sebakInfo.Node.Endpoint),
		false,
		http.Header{"Content-Type": []string{"application/json"}},
	)
	if err != nil {
		jw.WriteObject(err)
		return
	}

	var b []byte
	if b, err = client.Post("/api/v1/transactions", body, nil); err != nil {
		if se, ok := err.(*sebakerrors.Error); ok {
			w.WriteHeader(se.GetData("status").(int))
			w.Write([]byte(se.GetData("body").(string)))
			return
		}

		jw.WriteObject(err)
		return
	}

	w.Write(b)
}

func (h *Handler) GetTransactionByHash(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["id"]

	jw := NewJSONWriter(w)

	tx, err := item.GetTransaction(h.st, hash)
	if err != nil {
		jw.WriteObject(err)
		return
	}

	jw.WriteObject(resourcev1.NewTransaction(tx))
}

func (h *Handler) GetTransactionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["id"]

	jw := NewJSONWriter(w)

	status := "notfound"
	if found, err := item.ExistsTransaction(h.st, hash); err != nil {
		jw.WriteObject(err)
		return
	} else if found {
		status = "confirmed"
	} else {
		if found, _ := h.sst.Has(sebakblock.GetTransactionPoolKey(hash)); found {
			status = "submitted"
		}
	}

	if status == "notfound" {
		jw.WriteHeader(http.StatusNotFound)
		jw.WriteObject(sebakerrors.TransactionNotFound)
		return
	}

	if status == "submitted" {
		jw.WriteHeader(http.StatusAccepted)
	}

	payload := sebakresource.NewTransactionStatus(hash, status)
	jw.WriteObject(payload)
}