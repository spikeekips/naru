package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/naru/storage"
	"github.com/spikeekips/naru/storage/item"
)

type OperationsByAccountStreamHandler struct {
	H       *Handler
	w       http.ResponseWriter
	r       *http.Request
	address string
}

func (g OperationsByAccountStreamHandler) NewRequest(w http.ResponseWriter, r *http.Request) (StreamHandler, error) {
	vars := mux.Vars(r)
	address := vars["id"]

	if _, err := item.GetAccount(g.H.st, address); err != nil {
		return nil, err
	}

	return &OperationsByAccountStreamHandler{H: g.H, w: w, r: r, address: address}, nil
}

func (g *OperationsByAccountStreamHandler) Init() <-chan interface{} {
	iterFunc, closeFunc := item.GetOperationsByAccount(
		g.H.st,
		g.address,
		sebakstorage.NewDefaultListOptions(true, nil, 2),
	)

	ch := make(chan interface{})

	go func() {
		defer closeFunc()
		defer close(ch)

		for {
			op, hasNext, _ := iterFunc()
			if !hasNext {
				break
			}

			ch <- op
		}
		ch <- true
	}()

	return ch
}

func (g *OperationsByAccountStreamHandler) Stream() (<-chan interface{}, func()) {
	event := item.GetOperationAccountRelatedEventKey(g.address)

	ch := make(chan interface{})
	callback := func(items ...interface{}) {
		for _, v := range items {
			ch <- v
		}
	}

	storage.Observer.On(event, callback)

	return ch, func() {
		storage.Observer.Off(event, callback)
		close(ch)
	}
}
