package sebak

import (
	"fmt"

	sebakblock "boscoin.io/sebak/lib/block"
	sebakcommon "boscoin.io/sebak/lib/common"
	sebakerrors "boscoin.io/sebak/lib/errors"
	sebakstorage "boscoin.io/sebak/lib/storage"
	sebaktransaction "boscoin.io/sebak/lib/transaction"

	"github.com/spikeekips/naru/storage"
	"github.com/spikeekips/naru/storage/item"
)

func GetBlockByHeight(s *Storage, height uint64) (block sebakblock.Block, err error) {
	// get last block
	iterFunc, closeFunc := s.GetIterator(
		sebakcommon.BlockPrefixHeight,
		sebakstorage.NewDefaultListOptions(true, nil, 1),
	)
	it, _ := iterFunc()
	closeFunc()
	if len(it.Key) < 1 {
		err = sebakerrors.BlockNotFound
		return
	}

	var key string
	if _, err = s.Get(BlockHeightKey(height), &key); err != nil {
		return
	}

	if _, err = s.Get(BlockHashKey(key), &block); err != nil {
		log.Error("failed to get last block from sebak", "error", err)
		return
	}

	return
}

func GetLastBlock(s *Storage) (block sebakblock.Block, err error) {
	// get last block
	iterFunc, closeFunc := s.GetIterator(
		sebakcommon.BlockPrefixHeight,
		sebakstorage.NewDefaultListOptions(true, nil, 1),
	)
	it, _ := iterFunc()
	closeFunc()
	if len(it.Key) < 1 {
		err = sebakerrors.BlockNotFound
		return
	}

	var key string
	if err = storage.Deserialize(it.Value, &key); err != nil {
		log.Error("failed to parse last block key", "error", err)
		return
	}

	if _, err = s.Get(BlockHashKey(key), &block); err != nil {
		log.Error("failed to get last block from sebak", "error", err)
		return
	}

	return
}

// Blocks returns iterator function and close function. It acts like
// `GetIterator`, but instead of returning `sebakstorage.IterItem` it returns
// `sebakblock.Block`.
func GetBlocks(s *Storage, options sebakstorage.ListOptions) (func() (sebakblock.Block, bool), func()) {
	itf, cf := s.GetIterator(sebakcommon.BlockPrefixHeight, options)

	iterFunc := func() (sebakblock.Block, bool) {
		it, next := itf()
		if !next {
			return sebakblock.Block{}, next
		}

		var key string
		if err := storage.Deserialize(it.Value, &key); err != nil {
			log.Error("failed to deserialize block key", "error", err)
			return sebakblock.Block{}, false
		}

		var block sebakblock.Block
		var err error
		if _, err = s.Get(BlockHashKey(key), &block); err != nil {
			log.Error("failed to get block from sebak", "error", err)
			return sebakblock.Block{}, false
		}

		return block, true
	}

	return iterFunc, cf
}

func GetTransactions(s *Storage, hashes ...string) (txs []item.Transaction, err error) {
	for _, hash := range hashes {
		if len(hash) < 1 {
			continue
		}

		var tp sebakblock.TransactionPool
		if _, err = s.Get(TransactionKey(hash), &tp); err != nil {
			return
		}

		var tx sebaktransaction.Transaction
		if err = storage.Deserialize(tp.Message, &tx); err != nil {
			log.Error("failed to deserialize transaction message", "error", err)
			return
		}

		txs = append(txs, item.NewTransaction(tx, tp.Message))
	}

	return
}

func BlockHashKey(hash string) string {
	return fmt.Sprintf("%s%s", sebakcommon.BlockPrefixHash, hash)
}

func BlockHeightKey(height uint64) string {
	return fmt.Sprintf("%s%020d", sebakcommon.BlockPrefixHeight, height)
}

func TransactionKey(hash string) string {
	return fmt.Sprintf("%s%s", sebakcommon.TransactionPoolPrefix, hash)
}

func AccountKey(address string) string {
	return fmt.Sprintf("%s%s", sebakcommon.BlockAccountPrefixAddress, address)
}

func GetAccounts(s *Storage, options sebakstorage.ListOptions) (func() (item.Account, bool), func()) {
	itf, cf := s.GetIterator(sebakcommon.BlockAccountPrefixAddress, options)

	iterFunc := func() (item.Account, bool) {
		it, next := itf()
		if !next {
			return item.Account{}, next
		}

		var ac sebakblock.BlockAccount
		if err := storage.Deserialize(it.Value, &ac); err != nil {
			log.Error("failed to deserialize account", "error", err)
			return item.Account{}, false
		}

		return item.NewAccount(ac), true
	}

	return iterFunc, cf
}

func GetAccount(s *Storage, address string) (item.Account, error) {
	var ac sebakblock.BlockAccount
	if _, err := s.Get(AccountKey(address), &ac); err != nil {
		return item.Account{}, err
	}

	return item.NewAccount(ac), nil
}
