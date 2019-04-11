package sebak

import (
	"fmt"

	sebakblock "boscoin.io/sebak/lib/block"
	sebakcommon "boscoin.io/sebak/lib/common"
	sebaktransaction "boscoin.io/sebak/lib/transaction"

	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/storage"
)

func GetBlockByHeight(s *Storage, height uint64) (sebakblock.Block, error) {
	// get last block
	iterFunc, closeFunc, err := s.Iterator(
		sebakcommon.BlockPrefixHeight,
		storage.NewDefaultListOptions(true, nil, 1),
	)
	if err != nil {
		return sebakblock.Block{}, err
	}
	defer closeFunc()

	it, _, err := iterFunc()
	if err != nil {
		return sebakblock.Block{}, err
	}
	if len(it.Key) < 1 {
		return sebakblock.Block{}, BlockNotFound.New()
	}

	var key string
	if _, err := s.Get(BlockHeightKey(height), &key); err != nil {
		return sebakblock.Block{}, err
	}

	var block sebakblock.Block
	if _, err := s.Get(BlockHashKey(key), &block); err != nil {
		log.Error("failed to get last block from sebak", "error", err)
		return sebakblock.Block{}, err
	}

	return block, nil
}

func GetLastBlock(s *Storage) (sebakblock.Block, error) {
	// get last block
	iterFunc, closeFunc, err := s.Iterator(
		sebakcommon.BlockPrefixHeight,
		storage.NewDefaultListOptions(true, nil, 1),
	)
	if err != nil {
		return sebakblock.Block{}, err
	}
	defer closeFunc()

	it, _, err := iterFunc()
	if err != nil {
		return sebakblock.Block{}, err
	}

	if len(it.Key) < 1 {
		return sebakblock.Block{}, BlockNotFound.New()
	}

	var key string
	if err := storage.Deserialize(it.Value, &key); err != nil {
		log.Error("failed to parse last block key", "error", err)
		return sebakblock.Block{}, err
	}

	var block sebakblock.Block
	if _, err := s.Get(BlockHashKey(key), &block); err != nil {
		log.Error("failed to get last block from sebak", "error", err)
		return sebakblock.Block{}, err
	}

	return block, nil
}

// Blocks returns iterator function and close function. It acts like
// `Iterator`, but instead of returning `sebakstorage.IterItem` it returns
// `sebakblock.Block`.
func GetBlocks(s *Storage, options storage.ListOptions) (func() (sebakblock.Block, bool), func()) {
	itf, cf, err := s.Iterator(sebakcommon.BlockPrefixHeight, options)
	if err != nil {
		return nil, nil
	}

	iterFunc := func() (sebakblock.Block, bool) {
		it, next, err := itf()
		if err != nil || !next {
			return sebakblock.Block{}, next
		}

		var key string
		if err := storage.Deserialize(it.Value, &key); err != nil {
			log.Error("failed to deserialize block key", "error", err)
			return sebakblock.Block{}, false
		}

		var block sebakblock.Block
		if _, err := s.Get(BlockHashKey(key), &block); err != nil {
			log.Error("failed to get block from sebak", "error", err)
			return sebakblock.Block{}, false
		}

		return block, true
	}

	return iterFunc, cf
}

func GetTransactions(s *Storage, hashes ...string) (txs []element.TransactionMessage, err error) {
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

		txs = append(txs, element.NewTransactionMessage(tx, tp.Message))
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

func GetAccounts(s *Storage, options storage.ListOptions) (func() (element.Account, bool), func()) {
	itf, cf, err := s.Iterator(sebakcommon.BlockAccountPrefixAddress, options)
	if err != nil {
		return nil, nil
	}

	iterFunc := func() (element.Account, bool) {
		it, next, err := itf()
		if err != nil || !next {
			return element.Account{}, next
		}

		var ac sebakblock.BlockAccount
		if err := storage.Deserialize(it.Value, &ac); err != nil {
			log.Error("failed to deserialize account", "error", err)
			return element.Account{}, false
		}

		return element.NewAccount(ac), true
	}

	return iterFunc, cf
}

func GetAccount(s *Storage, address string) (element.Account, error) {
	var ac sebakblock.BlockAccount
	if _, err := s.Get(AccountKey(address), &ac); err != nil {
		return element.Account{}, err
	}

	return element.NewAccount(ac), nil
}
