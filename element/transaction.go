package element

import (
	"encoding/json"
	"fmt"
	"time"

	sebakblock "boscoin.io/sebak/lib/block"
	sebakcommon "boscoin.io/sebak/lib/common"
	sebaktransaction "boscoin.io/sebak/lib/transaction"
	sebakoperation "boscoin.io/sebak/lib/transaction/operation"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/storage"
)

type TransactionMessage struct {
	sebaktransaction.Transaction
	Raw []byte `json:"raw"`
}

func NewTransactionMessage(tx sebaktransaction.Transaction, raw []byte) TransactionMessage {
	return TransactionMessage{Transaction: tx, Raw: raw}
}

type Transaction struct {
	Raw []byte `json:"raw"`

	Hash       string             `json:"hash"`
	Block      uint64             `json:"block"`
	SequenceID uint64             `json:"sequence_id"`
	Signature  string             `json:"signature"`
	Source     string             `json:"source"`
	Fee        sebakcommon.Amount `json:"fee"`
	Operations []string           `json:"operations"`
	Amount     sebakcommon.Amount `json:"amount"`
	Confirmed  time.Time          `json:"confirmed"`
	Created    time.Time          `json:"created"`

	tx       sebaktransaction.Transaction
	block    Block
	accounts []string
}

func NewTransaction(tx sebaktransaction.Transaction, block Block, raw []byte) Transaction {
	var opHashes []string
	for _, op := range tx.B.Operations {
		opHashes = append(
			opHashes,
			sebakblock.NewBlockOperationKey(
				sebakcommon.MustMakeObjectHashString(op),
				tx.GetHash(),
			),
		)
	}

	created, _ := common.ParseISO8601(tx.H.Created)
	return Transaction{
		Hash:       tx.H.Hash,
		Block:      block.Header.Height,
		SequenceID: tx.B.SequenceID,
		Signature:  tx.H.Signature,
		Source:     tx.B.Source,
		Fee:        tx.B.Fee,
		Operations: opHashes,
		Amount:     tx.TotalAmount(true),
		Confirmed:  block.Header.ProposedTime,
		Created:    created,
		Raw:        raw,
		tx:         tx,
		block:      block,
	}
}

func (t Transaction) Save(st storage.Storage) error {
	if err := st.Insert(GetTransactionKey(t.Hash), t); err != nil {
		return err
	}

	st.Event("OnAfterSaveTransaction", st, t, t.tx, t.block)

	for opIndex, op := range t.tx.B.Operations {
		o, err := NewOperation(op, t.tx, uint64(opIndex), t.block.Header.Height)
		if err != nil {
			return err
		}
		if err := o.Save(st); err != nil {
			return err
		}
	}

	return nil
}

func (t Transaction) Transaction() (tx sebaktransaction.Transaction, err error) {
	err = json.Unmarshal(t.Raw, &tx)
	return
}

func (t Transaction) AllAccounts() []string {
	if t.accounts != nil {
		return t.accounts
	}

	var isProposerTransaction bool
	addresses := map[string]struct{}{}
	for _, op := range t.tx.B.Operations {
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
		t.accounts = append(t.accounts, t.Source)
	}

	for address, _ := range addresses {
		t.accounts = append(t.accounts, address)
	}

	return t.accounts
}

func GetTransactionKey(hash string) string {
	return fmt.Sprintf("%s%s", TransactionPrefix, hash)
}
