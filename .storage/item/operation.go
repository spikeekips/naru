package item

import (
	"fmt"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebakstorage "boscoin.io/sebak/lib/storage"
	sebaktransaction "boscoin.io/sebak/lib/transaction"
	sebakoperation "boscoin.io/sebak/lib/transaction/operation"

	"github.com/spikeekips/naru/storage"
)

type Operation struct {
	Hash string `json:"hash"`

	OpHash     string                       `json:"op_hash"`
	OpIndex    uint64                       `json:"op_index"`
	TxHash     string                       `json:"tx_hash"`
	Type       sebakoperation.OperationType `json:"type"`
	Source     string                       `json:"source"`
	Target     string                       `json:"target"`
	Height     uint64                       `json:"block_height"`
	SequenceID uint64                       `json:"sequence_id"`
	Linked     string                       `json:"linked"`
	Raw        []byte                       `json:"raw"`
}

func NewOperation(op sebakoperation.Operation, tx sebaktransaction.Transaction, opIndex uint64, blockHeight uint64) (Operation, error) {
	opHash := sebakcommon.MustMakeObjectHashString(op)
	txHash := tx.GetHash()

	target := ""
	if pop, ok := op.B.(sebakoperation.Targetable); ok {
		target = pop.TargetAddress()
	}

	linked := ""
	if createAccount, ok := op.B.(sebakoperation.CreateAccount); ok {
		if createAccount.Linked != "" {
			linked = createAccount.Linked
		}
	}

	raw := sebakcommon.MustMarshalJSON(op)

	return Operation{
		Hash: GetOperationHash(txHash, opIndex),

		OpHash:     opHash,
		OpIndex:    opIndex,
		TxHash:     txHash,
		Type:       op.H.Type,
		Source:     tx.B.Source,
		Target:     target,
		Height:     blockHeight,
		SequenceID: tx.B.SequenceID,
		Linked:     linked,
		Raw:        raw,
	}, nil
}

func GetOperationHash(txHash string, opIndex uint64) string {
	return fmt.Sprintf("%s%06d", txHash, opIndex)
}
func GetOperationKey(hash string) string {
	return fmt.Sprintf("%s%s", storage.OperationPrefix, hash)
}

func GetOperationAccountRelatedEventKey(address string) string {
	return fmt.Sprintf(
		"%s%s%s",
		storage.EventNewItem,
		storage.OperationAccountRelatedPrefix,
		address,
	)
}

func GetOperationAccountRelatedKey(address string, blockHeight uint64) string {
	return fmt.Sprintf(
		"%s%s%20d%s",
		storage.OperationAccountRelatedPrefix,
		address,
		blockHeight,
		sebakcommon.GetUniqueIDFromUUID(),
	)
}

func (o Operation) Save(st *storage.Storage) error {
	if err := st.New(GetOperationKey(o.Hash), o); err != nil {
		return err
	}
	if err := st.New(GetOperationAccountRelatedKey(o.Source, o.Height), o.Hash); err != nil {
		return err
	}

	event := GetOperationAccountRelatedEventKey(o.Source)
	if len(o.Target) > 0 {
		if err := st.New(GetOperationAccountRelatedKey(o.Target, o.Height), o.Hash); err != nil {
			return err
		}

		event += " " + GetOperationAccountRelatedEventKey(o.Target)
	}

	st.AddEvent(event, o)

	return nil
}

func GetOperation(st *storage.Storage, hash string) (op Operation, err error) {
	err = st.Get(GetOperationKey(hash), &op)
	return
}

func GetOperationsByAccount(st *storage.Storage, address string, options sebakstorage.ListOptions) (
	func() (Operation, bool, []byte),
	func(),
) {
	iterFunc, closeFunc := st.GetIterator(fmt.Sprintf("%s%s", storage.OperationAccountRelatedPrefix, address), options)

	return (func() (Operation, bool, []byte) {
			item, hasNext := iterFunc()
			if !hasNext {
				return Operation{}, false, item.Key
			}

			var hash string
			sebakcommon.MustUnmarshalJSON(item.Value, &hash)

			o, err := GetOperation(st, hash)
			if err != nil {
				return Operation{}, false, item.Key
			}

			return o, hasNext, item.Key
		}), (func() {
			closeFunc()
		})
}
