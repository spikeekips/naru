package resourcev1

import (
	"strings"

	"github.com/nvellon/hal"

	sebakresource "boscoin.io/sebak/lib/node/runner/api/resource"

	"github.com/spikeekips/naru/element"
)

type Transaction struct {
	tx element.Transaction
}

func NewTransaction(tx element.Transaction) *Transaction {
	t := &Transaction{
		tx: tx,
	}
	return t
}

func (t Transaction) GetMap() hal.Entry {
	tx, err := t.tx.Transaction()
	if err != nil {
		panic(err)
	}

	return hal.Entry{
		"hash":            t.tx.Hash,
		"block":           t.tx.Block,
		"source":          t.tx.Source,
		"fee":             t.tx.Fee.String(),
		"sequence_id":     t.tx.SequenceID,
		"created":         t.tx.Created,
		"operation_count": len(t.tx.Operations),
		"operations":      tx.B.Operations,
	}
}
func (t Transaction) Resource() *hal.Resource {
	r := hal.NewResource(t, t.LinkSelf())
	r.AddLink("account", hal.NewLink(strings.Replace(sebakresource.URLAccounts, "{id}", t.tx.Source, -1)))
	r.AddLink(
		"operations",
		hal.NewLink(
			strings.Replace(
				sebakresource.URLTransactionOperations,
				"{id}",
				t.tx.Hash, -1,
			)+"{?cursor,limit,order}",
			hal.LinkAttr{"templated": true},
		),
	)
	return r
}

func (t Transaction) LinkSelf() string {
	// TODO support FQDN
	return strings.Replace(sebakresource.URLTransactions, "{id}", t.tx.Hash, -1)
}
