package mongoelement

import (
	"strings"

	sebaktransaction "boscoin.io/sebak/lib/transaction"
	sebakoperation "boscoin.io/sebak/lib/transaction/operation"

	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/storage"
)

func EventSync() {
	storage.Observer.Sync("OnAfterSaveAccount", OnAfterSaveAccount)
	storage.Observer.Sync("OnAfterSaveTransaction", OnAfterSaveTransaction)
	storage.Observer.Sync("OnAfterSaveOperation", OnAfterSaveOperation)
	storage.Observer.Sync("OnAfterDigest", OnAfterDigest)
}

func OnAfterSaveAccount(st storage.Storage, account element.Account, created bool) {
}

func OnAfterDigest(potion element.Potion, start, end uint64) {
	iterFunc, closeFunc := potion.OperationsByHeight(start, end+1)
	defer closeFunc()

	var createdAccounts []element.Operation
	var added uint64
	for {
		operation, next, _ := iterFunc()
		if !next {
			break
		}
		switch operation.Type {
		case sebakoperation.TypeCreateAccount:
			createdAccounts = append(createdAccounts, operation)
		case sebakoperation.TypeInflation, sebakoperation.TypeInflationPF:
		default:
			continue
		}

		added += uint64(operation.Amount)
	}

	{ // BlockStat
		bs, err := potion.BlockStat()
		if err != nil {
			if !storage.NotFound.Equal(err) {
				log.Error("failed to get BlockStat", "error", err)
				return
			}
			bs = element.NewBlockStat()
		}

		before := bs.TotalSupply

		bs.TotalSupply += added
		if err := bs.Save(potion.Storage()); err != nil {
			log.Error("failed to save BlockStat", "error", err)
			return
		}
		log.Debug(
			"BlockStat.totalSupply updated",
			"block-start", start,
			"block-end", end,
			"before", before,
			"after", bs.TotalSupply,
			"added", added,
		)
	}

	{ // Account.Created
		for _, operation := range createdAccounts {
			ac, err := potion.Account(operation.Target)
			if err != nil {
				log.Error("failed to get account", "error", err)
				continue
			}
			ac.CreatedHeight = operation.Height

			if err := potion.Storage().Update(element.GetAccountKey(ac.Address), ac); err != nil {
				log.Error("failed to save account with CreatedHeight", "error", err)
				continue
			}
		}
	}
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
