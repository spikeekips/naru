package item

import (
	"fmt"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebaktransaction "boscoin.io/sebak/lib/transaction"
	sebakoperation "boscoin.io/sebak/lib/transaction/operation"

	storage "github.com/spikeekips/naru/newstorage"
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
	return fmt.Sprintf("%s%s", OperationPrefix, hash)
}

func GetOperationAccountRelatedEventKey(address string) string {
	return fmt.Sprintf(
		"%s%s%s",
		EventPrefixNewOperation,
		OperationAccountRelatedPrefix,
		address,
	)
}

func GetOperationAccountRelatedKey(address string, blockHeight uint64) string {
	return fmt.Sprintf(
		"%s%s%20d%s",
		OperationAccountRelatedPrefix,
		address,
		blockHeight,
		sebakcommon.GetUniqueIDFromUUID(),
	)
}

func (o Operation) Save(st storage.Storage) error {
	if err := st.Insert(GetOperationKey(o.Hash), o); err != nil {
		return err
	}

	st.Event("OnAfterSaveOperation", st, o)

	return nil
}

func GetOperation(st storage.Storage, hash string) (op Operation, err error) {
	err = st.Get(GetOperationKey(hash), &op)
	return
}

func GetOperationsByAccount(st storage.Storage, address string, options storage.ListOptions) (
	func() (Operation, bool, []byte),
	func(),
) {
	iterFunc, closeFunc := st.Iterator(fmt.Sprintf("%s%s", OperationAccountRelatedPrefix, address), "", options)

	return (func() (Operation, bool, []byte) {
			item, hasNext := iterFunc()
			if !hasNext {
				return Operation{}, false, []byte(item.Key)
			}

			hash, ok := item.Value.(string)
			if !ok {
				return Operation{}, false, []byte(item.Key)
			}

			o, err := GetOperation(st, hash)
			if err != nil {
				return Operation{}, false, []byte(item.Key)
			}

			return o, hasNext, []byte(item.Key)
		}), (func() {
			closeFunc()
		})
}
