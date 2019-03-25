package mongoitem

import (
	"fmt"

	sebakerrors "boscoin.io/sebak/lib/errors"
	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/naru/storage"
	mongostorage "github.com/spikeekips/naru/storage/backend/mongo"
	"github.com/spikeekips/naru/storage/item"
)

type Getter struct {
	s *mongostorage.Storage
}

func NewGetter(s *mongostorage.Storage) Getter {
	return Getter{s: s}
}

func (g Getter) Storage() storage.Storage {
	return g.s
}

func (g Getter) Account(address string) (item.Account, error) {
	var ac item.Account
	err := g.s.Get(item.GetAccountKey(address), &ac)
	return ac, err
}

func (g Getter) Block(hash string) (item.Block, error) {
	var block item.Block
	if err := g.s.Get(item.GetBlockKey(hash), &block); err != nil {
		return item.Block{}, err
	}

	return block, nil
}

func (g Getter) BlockByHeight(height uint64) (item.Block, error) {
	var hash string
	if err := g.s.Get(item.GetBlockHeightKey(height), &hash); err != nil {
		return item.Block{}, err
	}

	return g.Block(hash)
}

func (g Getter) LastBlock() (item.Block, error) {
	options := storage.NewDefaultListOptions(true, nil, 1)
	iterFunc, closeFunc := g.s.Iterator(item.BlockHeightPrefix, "", options)
	i, hasNext := iterFunc()
	closeFunc()
	if !hasNext {
		return item.Block{}, sebakerrors.StorageRecordDoesNotExist
	}

	hash, ok := i.Value.(string)
	if !ok {
		return item.Block{}, storage.DecodeValueFailed.New()
	}

	return g.Block(hash)
}

func (g Getter) BlocksIterator(
	iterFunc func() (sebakstorage.IterItem, bool),
	closeFunc func(),
) (
	func() (item.Block, bool, []byte),
	func(),
) {

	return (func() (item.Block, bool, []byte) {
			it, hasNext := iterFunc()
			if !hasNext {
				return item.Block{}, false, []byte{}
			}

			var hash string
			if err := storage.Deserialize(it.Value, &hash); err != nil {
				return item.Block{}, false, []byte{}
			}

			b, err := g.Block(hash)
			if err != nil {
				return item.Block{}, false, []byte{}
			}

			return b, hasNext, it.Key
		}), (func() {
			closeFunc()
		})
}

func (g Getter) Operation(hash string) (op item.Operation, err error) {
	err = g.s.Get(item.GetOperationKey(hash), &op)
	return
}

func (g Getter) OperationsByAccount(address string, options storage.ListOptions) (
	func() (item.Operation, bool, []byte),
	func(),
) {
	iterFunc, closeFunc := g.s.Iterator(fmt.Sprintf("%s%s", item.OperationAccountRelatedPrefix, address), "", options)

	return (func() (item.Operation, bool, []byte) {
			it, hasNext := iterFunc()
			if !hasNext {
				return item.Operation{}, false, []byte(it.Key)
			}

			hash, ok := it.Value.(string)
			if !ok {
				return item.Operation{}, false, []byte(it.Key)
			}

			o, err := g.Operation(hash)
			if err != nil {
				return item.Operation{}, false, []byte(it.Key)
			}

			return o, hasNext, []byte(it.Key)
		}), (func() {
			closeFunc()
		})
}

func (g Getter) ExistsTransaction(hash string) (bool, error) {
	return g.s.Has(item.GetTransactionKey(hash))
}

func (g Getter) Transaction(hash string) (tx item.Transaction, err error) {
	if err = g.s.Get(item.GetTransactionKey(hash), &tx); err != nil {
		return
	}
	return
}
