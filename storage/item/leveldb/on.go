package leveldbitem

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
	if err := st.Insert(item.GetBlockHeightKey(block.Height), block.Hash); err != nil {
		return
	}
}

func OnAfterSaveTransaction(st storage.Storage, transaction item.Transaction, tx sebaktransaction.Transaction, block item.Block) {
	if err := st.Insert(item.GetTransactionBlockKey(block.Height), transaction.Hash); err != nil {
		return
	}
	if err := st.Insert(item.GetTransactionSourceKey(transaction.Source, block.Height), transaction.Hash); err != nil {
		return
	}

	for _, address := range transaction.AllAccounts() {
		if err := st.Insert(item.GetTransactionAccountsKey(address, block.Height), transaction.Hash); err != nil {
			return
		}
	}
}

func OnAfterSaveOperation(st storage.Storage, operation item.Operation) {
	if err := st.Insert(item.GetOperationAccountRelatedKey(operation.Source, operation.Height), operation.Hash); err != nil {
		return
	}

	var events []string = []string{item.GetOperationAccountRelatedEventKey(operation.Source)}
	if len(operation.Target) > 0 {
		if err := st.Insert(item.GetOperationAccountRelatedKey(operation.Target, operation.Height), operation.Hash); err != nil {
			return
		}

		events = append(events, item.GetOperationAccountRelatedEventKey(operation.Target))
	}

	st.Event(strings.Join(events, " "), operation)
}
