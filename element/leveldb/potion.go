package leveldbelement

import (
	"fmt"

	sebakerrors "boscoin.io/sebak/lib/errors"
	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/storage"
	leveldbstorage "github.com/spikeekips/naru/storage/backend/leveldb"
)

func GetBlockHeightKey(height uint64) string {
	return fmt.Sprintf("%s%020d", BlockHeightPrefix, height)
}

func GetOperationAccountRelatedKey(address string, blockHeight uint64) string {
	return fmt.Sprintf(
		"%s%s%20d%s",
		element.OperationAccountRelatedPrefix,
		address,
		blockHeight,
		common.SequentialUUID(),
	)
}

func GetTransactionBlockKey(block uint64) string {
	return fmt.Sprintf(
		"%s%020d%s",
		TransactionBlockPrefix,
		block,
		common.SequentialUUID(),
	)
}

func GetTransactionSourceKey(source string, block uint64) string {
	return fmt.Sprintf(
		"%s%s%020d%s",
		TransactionSourcePrefix,
		source,
		block,
		common.SequentialUUID(),
	)
}

func GetTransactionAccountsKey(address string, block uint64) string {
	return fmt.Sprintf(
		"%s%s%020d%s",
		TransactionAccountsPrefix,
		address,
		block,
		common.SequentialUUID(),
	)
}

type Potion struct {
	s *leveldbstorage.Storage
}

func NewPotion(s *leveldbstorage.Storage) Potion {
	return Potion{s: s}
}

func (g Potion) Storage() storage.Storage {
	return g.s
}

func (g Potion) Account(address string) (element.Account, error) {
	var ac element.Account
	err := g.s.Get(element.GetAccountKey(address), &ac)
	return ac, err
}

func (g Potion) Block(hash string) (element.Block, error) {
	var block element.Block
	if err := g.s.Get(element.GetBlockKey(hash), &block); err != nil {
		return element.Block{}, err
	}

	return block, nil
}

func (g Potion) BlockByHeight(height uint64) (element.Block, error) {
	var hash string
	if err := g.s.Get(GetBlockHeightKey(height), &hash); err != nil {
		return element.Block{}, err
	}

	return g.Block(hash)
}

func (g Potion) LastBlock() (element.Block, error) {
	options := storage.NewDefaultListOptions(true, nil, 1)
	iterFunc, closeFunc := g.s.Iterator(BlockHeightPrefix, "", options)
	i, hasNext := iterFunc()
	closeFunc()
	if !hasNext {
		return element.Block{}, sebakerrors.StorageRecordDoesNotExist
	}

	hash, ok := i.Value.(string)
	if !ok {
		return element.Block{}, storage.DecodeValueFailed.New()
	}

	return g.Block(hash)
}

func (g Potion) BlocksIterator(
	iterFunc func() (sebakstorage.IterItem, bool),
	closeFunc func(),
) (
	func() (element.Block, bool, []byte),
	func(),
) {

	return (func() (element.Block, bool, []byte) {
			it, hasNext := iterFunc()
			if !hasNext {
				return element.Block{}, false, []byte{}
			}

			var hash string
			if err := storage.Deserialize(it.Value, &hash); err != nil {
				return element.Block{}, false, []byte{}
			}

			b, err := g.Block(hash)
			if err != nil {
				return element.Block{}, false, []byte{}
			}

			return b, hasNext, it.Key
		}), (func() {
			closeFunc()
		})
}

func (g Potion) Operation(hash string) (op element.Operation, err error) {
	err = g.s.Get(element.GetOperationKey(hash), &op)
	return
}

func (g Potion) OperationsByAccount(address string, options storage.ListOptions) (
	func() (element.Operation, bool, []byte),
	func(),
) {
	iterFunc, closeFunc := g.s.Iterator(fmt.Sprintf("%s%s", element.OperationAccountRelatedPrefix, address), "", options)

	return (func() (element.Operation, bool, []byte) {
			it, hasNext := iterFunc()
			if !hasNext {
				return element.Operation{}, false, []byte(it.Key)
			}

			hash, ok := it.Value.(string)
			if !ok {
				return element.Operation{}, false, []byte(it.Key)
			}

			o, err := g.Operation(hash)
			if err != nil {
				return element.Operation{}, false, []byte(it.Key)
			}

			return o, hasNext, []byte(it.Key)
		}), (func() {
			closeFunc()
		})
}

func (g Potion) ExistsTransaction(hash string) (bool, error) {
	return g.s.Has(element.GetTransactionKey(hash))
}

func (g Potion) Transaction(hash string) (tx element.Transaction, err error) {
	if err = g.s.Get(element.GetTransactionKey(hash), &tx); err != nil {
		return
	}
	return
}
