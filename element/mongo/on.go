package mongoelement

import (
	"strings"

	sebaktransaction "boscoin.io/sebak/lib/transaction"

	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/storage"
)

func EventSync() {
	storage.Observer.Sync("OnAfterSaveAccount", OnAfterSaveAccount)
	storage.Observer.Sync("OnAfterSaveBlock", OnAfterSaveBlock)
	storage.Observer.Sync("OnAfterSaveTransaction", OnAfterSaveTransaction)
	storage.Observer.Sync("OnAfterSaveOperation", OnAfterSaveOperation)
}

func OnAfterSaveAccount(st storage.Storage, account element.Account) {
}

func OnAfterSaveBlock(st storage.Storage, block element.Block) {
}

func OnAfterSaveTransaction(st storage.Storage, transaction element.Transaction, tx sebaktransaction.Transaction, block element.Block) {
}

func OnAfterSaveOperation(st storage.Storage, operation element.Operation) {
	var events []string = []string{element.GetOperationAccountRelatedEventKey(operation.Source)}
	if len(operation.Target) > 0 {
		events = append(events, element.GetOperationAccountRelatedEventKey(operation.Target))
	}

	st.Event(strings.Join(events, " "), operation)
}
