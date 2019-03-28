package leveldbelement

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
	if err := st.Insert(GetBlockHeightKey(block.Height), block.Hash); err != nil {
		return
	}
}

func OnAfterSaveTransaction(st storage.Storage, transaction element.Transaction, tx sebaktransaction.Transaction, block element.Block) {
	if err := st.Insert(GetTransactionBlockKey(block.Height), transaction.Hash); err != nil {
		return
	}
	if err := st.Insert(GetTransactionSourceKey(transaction.Source, block.Height), transaction.Hash); err != nil {
		return
	}

	for _, address := range transaction.AllAccounts() {
		if err := st.Insert(GetTransactionAccountsKey(address, block.Height), transaction.Hash); err != nil {
			return
		}
	}
}

func OnAfterSaveOperation(st storage.Storage, operation element.Operation) {
	if err := st.Insert(GetOperationAccountRelatedKey(operation.Source, operation.Height), operation.Hash); err != nil {
		return
	}

	var events []string = []string{element.GetOperationAccountRelatedEventKey(operation.Source)}
	if len(operation.Target) > 0 {
		if err := st.Insert(GetOperationAccountRelatedKey(operation.Target, operation.Height), operation.Hash); err != nil {
			return
		}

		events = append(events, element.GetOperationAccountRelatedEventKey(operation.Target))
	}

	st.Event(strings.Join(events, " "), operation)
}
