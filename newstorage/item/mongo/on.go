package mongoitem

import (
	"strings"

	sebaktransaction "boscoin.io/sebak/lib/transaction"

	"github.com/spikeekips/naru/newstorage"
	"github.com/spikeekips/naru/newstorage/item"
)

func EventSync() {
	newstorage.Observer.Sync("OnAfterSaveAccount", OnAfterSaveAccount)
	newstorage.Observer.Sync("OnAfterSaveBlock", OnAfterSaveBlock)
	newstorage.Observer.Sync("OnAfterSaveTransaction", OnAfterSaveTransaction)
	newstorage.Observer.Sync("OnAfterSaveOperation", OnAfterSaveOperation)
}

func OnAfterSaveAccount(st newstorage.Storage, account item.Account) {
}

func OnAfterSaveBlock(st newstorage.Storage, block item.Block) {
	if err := st.Insert(item.GetBlockHeightKey(block.Height), block.Hash); err != nil {
		return
	}
}

func OnAfterSaveTransaction(st newstorage.Storage, transaction item.Transaction, tx sebaktransaction.Transaction, block item.Block) {
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

func OnAfterSaveOperation(st newstorage.Storage, operation item.Operation) {
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
