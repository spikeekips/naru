package restv1

import (
	sebakapi "boscoin.io/sebak/lib/node/runner/api"
	"github.com/gorilla/mux"

	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/storage"
)

type OperationsByAccountStreamHandler struct {
	BaseStreamHandler
	H       *Handler
	address string
	query   *sebakapi.PageQuery
}

func (g OperationsByAccountStreamHandler) NewRequest(base BaseStreamHandler) (StreamHandler, error) {
	vars := mux.Vars(base.Request())
	address := vars["id"]

	if _, err := g.H.potion.Account(address); err != nil {
		return nil, err
	}

	query, err := sebakapi.NewPageQuery(base.Request())
	if err != nil {
		return nil, err
	}

	return &OperationsByAccountStreamHandler{BaseStreamHandler: base, H: g.H, address: address, query: query}, nil
}

func (g *OperationsByAccountStreamHandler) Init() <-chan interface{} {
	// TODO
	lo := g.query.ListOptions()
	iterFunc, closeFunc := g.H.potion.OperationsByAccount(
		g.address,
		storage.NewDefaultListOptions(lo.Reverse(), lo.Cursor(), lo.Limit()),
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
	event := element.GetOperationAccountRelatedEventKey(g.address)

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
