package item

import (
	"fmt"

	sebakblock "boscoin.io/sebak/lib/block"
	sebakerrors "boscoin.io/sebak/lib/errors"
	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/naru/storage"
)

type Block struct {
	sebakblock.Block
}

func NewBlock(block sebakblock.Block) Block {
	return Block{Block: block}
}

func (b Block) Save(st *storage.Storage) error {
	if err := st.New(GetBlockKey(b.Hash), b); err != nil {
		return err
	}

	if err := st.New(GetBlockHeightKey(b.Height), b.Hash); err != nil {
		return err
	}

	st.AddEvent(storage.EventNewBlock, b)

	return nil
}

func GetBlockKey(hash string) string {
	return fmt.Sprintf("%s%s", storage.BlockPrefix, hash)
}

func GetBlockHeightKey(height uint64) string {
	return fmt.Sprintf("%s%020d", storage.BlockHeightPrefix, height)
}

func GetBlock(st *storage.Storage, hash string) (Block, error) {
	var block Block
	if err := st.Get(GetBlockKey(hash), &block); err != nil {
		return Block{}, err
	}

	return block, nil
}

func GetBlockByHeight(st *storage.Storage, height uint64) (Block, error) {
	var hash string
	if err := st.Get(GetBlockHeightKey(height), &hash); err != nil {
		return Block{}, err
	}

	return GetBlock(st, hash)
}

func GetLastBlock(st *storage.Storage) (Block, error) {
	options := sebakstorage.NewDefaultListOptions(true, nil, 1)
	iterFunc, closeFunc := st.GetIterator(storage.BlockHeightPrefix, options)
	i, hasNext := iterFunc()
	closeFunc()
	if !hasNext {
		return Block{}, sebakerrors.StorageRecordDoesNotExist
	}

	var hash string
	if err := storage.Deserialize(i.Value, &hash); err != nil {
		return Block{}, err
	}

	return GetBlock(st, hash)
}

func BlocksIterator(
	st *storage.Storage,
	iterFunc func() (sebakstorage.IterItem, bool),
	closeFunc func(),
) (
	func() (Block, bool, []byte),
	func(),
) {

	return (func() (Block, bool, []byte) {
			item, hasNext := iterFunc()
			if !hasNext {
				return Block{}, false, []byte{}
			}

			var hash string
			if err := storage.Deserialize(item.Value, &hash); err != nil {
				return Block{}, false, []byte{}
			}

			b, err := GetBlock(st, hash)
			if err != nil {
				return Block{}, false, []byte{}
			}

			return b, hasNext, item.Key
		}), (func() {
			closeFunc()
		})
}
