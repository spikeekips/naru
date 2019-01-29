package item

import (
	"fmt"

	sebaktransaction "boscoin.io/sebak/lib/transaction"
	sebakoperation "boscoin.io/sebak/lib/transaction/operation"

	"github.com/spikeekips/naru/storage"
)

type Transaction struct {
	sebaktransaction.Transaction
	Raw      []byte `json:"raw"`
	accounts []string
}

func NewTransaction(tx sebaktransaction.Transaction, raw []byte) Transaction {
	return Transaction{Transaction: tx, Raw: raw}
}

func (t Transaction) Save(st *storage.Storage) error {
	return st.New(GetTransactionKey(t.GetHash()), t)
}

func (t Transaction) AllAccounts() []string {
	if t.accounts != nil {
		return t.accounts
	}

	var isProposerTransaction bool
	addresses := map[string]struct{}{}
	for _, op := range t.B.Operations {
		if top, ok := op.B.(sebakoperation.Targetable); !ok {
			continue
		} else {
			if !isProposerTransaction {
				if _, ok := op.B.(sebakoperation.CollectTxFee); ok {
					isProposerTransaction = true
				}
			}
			addresses[top.TargetAddress()] = struct{}{}
		}
	}
	if !isProposerTransaction {
		t.accounts = append(t.accounts, t.Source())
	}

	for address, _ := range addresses {
		t.accounts = append(t.accounts, address)
	}

	return t.accounts
}

func GetTransactionKey(hash string) string {
	return fmt.Sprintf("%s%s", storage.TransactionPrefix, hash)
}
