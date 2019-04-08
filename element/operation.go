package element

import (
	"fmt"

	sebakcommon "boscoin.io/sebak/lib/common"
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
	Amount     sebakcommon.Amount           `json:"amount"`
	Raw        []byte                       `json:"raw"`
}

func NewOperation(op sebakoperation.Operation, tx sebaktransaction.Transaction, opIndex uint64, blockHeight uint64) (Operation, error) {
	opHash := sebakcommon.MustMakeObjectHashString(op)
	txHash := tx.GetHash()

	target := ""
	if pop, ok := op.B.(sebakoperation.Targetable); ok {
		target = pop.TargetAddress()
	}

	amount := sebakcommon.Amount(0)
	if payable, ok := op.B.(sebakoperation.Payable); ok {
		amount = payable.GetAmount()
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
		Amount:     amount,
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

func (o Operation) Save(st storage.Storage) error {
	if err := st.Insert(GetOperationKey(o.Hash), o); err != nil {
		return err
	}

	st.Event("OnAfterSaveOperation", st, o)

	return nil
}
