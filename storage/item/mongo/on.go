package mongoitem

import (
	"strings"

	sebaktransaction "boscoin.io/sebak/lib/transaction"

	"github.com/spikeekips/naru/storage"
	"github.com/spikeekips/naru/storage/item"
)

func EventSync() {
	storage.Observer.Sync("OnAfterSaveAccount", OnAfterSaveAccount)
	storage.Observer.Sync("OnAfterSaveBlock", OnAfterSaveBlock)
	storage.Observer.Sync("OnAfterSaveTransaction", OnAfterSaveTransaction)
	storage.Observer.Sync("OnAfterSaveOperation", OnAfterSaveOperation)
}

func OnAfterSaveAccount(st storage.Storage, account item.Account) {
}

func OnAfterSaveBlock(st storage.Storage, block item.Block) {
}

func OnAfterSaveTransaction(st storage.Storage, transaction item.Transaction, tx sebaktransaction.Transaction, block item.Block) {
}

func OnAfterSaveOperation(st storage.Storage, operation item.Operation) {
	var events []string = []string{item.GetOperationAccountRelatedEventKey(operation.Source)}
	if len(operation.Target) > 0 {
		events = append(events, item.GetOperationAccountRelatedEventKey(operation.Target))
	}

	st.Event(strings.Join(events, " "), operation)
}
