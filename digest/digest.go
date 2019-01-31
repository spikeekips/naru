package digest

import (
	"time"

	logging "github.com/inconshreveable/log15"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebakerrors "boscoin.io/sebak/lib/errors"
	sebakrunner "boscoin.io/sebak/lib/node/runner"
	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
	"github.com/spikeekips/naru/storage/item"
)

var maxNumberOfWorkers int = 100

type Digest struct {
	st            *storage.Storage
	sst           *sebak.Storage
	genesisSource string
	blocksLimit   uint64
	start         uint64
	end           uint64
	initialize    bool
}

func NewDigest(st *storage.Storage, sst *sebak.Storage, genesisSource string, start, end uint64, initialize bool) *Digest {
	return &Digest{
		st:            st,
		sst:           sst.New(),
		genesisSource: genesisSource,
		start:         start,
		end:           end,
		blocksLimit:   sebakrunner.MaxLimitListOptions,
		initialize:    initialize,
	}
}

func (d *Digest) Open() error {
	err := d.sst.Provider().Open()
	if err == nil {
		log.Debug("sebak storage opened")
		return nil
	}

	if err == sebak.ProviderNotClosedError {
		log.Warn("failed to open sebak provider", "error", err)
		return nil
	}
	return err
}

func (d *Digest) Close() error {
	err := d.sst.Provider().Close()
	if err == nil {
		log.Debug("sebak storage closed")
		return nil
	}

	if err == sebak.ProviderNotOpenedError {
		log.Warn("failed to close sebak provider", "error", err)
		return nil
	}
	return err
}

func (d *Digest) Digest() error {
	numberOfWorkers := int((d.end - d.start) / d.blocksLimit)
	if numberOfWorkers < 1 {
		numberOfWorkers = 1
	} else if numberOfWorkers > maxNumberOfWorkers {
		numberOfWorkers = maxNumberOfWorkers
	}

	started := time.Now()
	log_ := log.New(logging.Ctx{
		"start":              d.start,
		"end":                d.end,
		"blocksLimit":        d.blocksLimit,
		"numberOfWorkers":    numberOfWorkers,
		"maxNumberOfWorkers": maxNumberOfWorkers,
		"started":            started,
	})

	log_.Debug("start digest")

	chanWorker := make(chan [2]uint64, 1000)
	chanError := make(chan error, 1000)

	for wid := 1; wid <= numberOfWorkers; wid++ {
		go d.digestBlocks(wid, chanWorker, chanError)
	}

	var countCursors int
	var cursorSet [][2]uint64
	var start uint64 = d.start
	for {
		if start >= d.end {
			break
		}
		end := start + d.blocksLimit
		if end > d.end {
			end = d.end
		}

		log_.Debug("cursor", "start", start, "end", end)
		cursorSet = append(cursorSet, [2]uint64{start, end})
		start += d.blocksLimit
		countCursors += 1
	}

	go func() {
		for _, cursors := range cursorSet {
			log_.Debug("cursors", "start", cursors[0], "end", cursors[1])
			chanWorker <- cursors
		}
		close(chanWorker)
	}()

	var count int
	var err error
end:
	for {
		select {
		case err = <-chanError:
			if err != nil {
				log_.Error("failed to digest block", "error", err)
				break end
			}

			count += 1
			if count == countCursors {
				break end
			}
		}
	}

	if err != nil {
		return err
	}

	if d.initialize {
		if err := d.saveAccounts(d.st); err != nil {
			return err
		}
	}

	d.logInsertedData()

	ended := time.Now()
	log_.Debug("digest done", "end", ended, "elapsed", ended.Sub(started))

	return nil
}

func (d *Digest) digestBlocks(wid int, chanWorker <-chan [2]uint64, chanError chan<- error) {
	for cursor := range chanWorker {
		err := d.digestBlocksByHeight(cursor[0], cursor[1])
		chanError <- err

		if err != nil {
			log.Error("failed digestBlocksByHeight", "error", err, "wid", wid, "cursor", cursor)
			break
		}
	}
}

func (d *Digest) digestBlocksByHeight(start, end uint64) error {
	log_ := log.New(logging.Ctx{
		"start": start,
		"end":   end,
	})

	log_.Debug("start digestBlocksByHeight")
	var cursor []byte
	if start < sebakcommon.GenesisBlockHeight {
		cursor = nil
	} else {
		cursor = []byte(sebak.BlockHeightKey(start))
	}

	options := sebakstorage.NewDefaultListOptions(
		false,
		cursor,
		d.blocksLimit,
	)
	iterFunc, closeFunc := sebak.GetBlocks(d.sst, options)
	defer closeFunc()

	var block item.Block
	var n int
	for {
		blk, hasNext := iterFunc()
		if !hasNext {
			break
		}
		if blk.Height > end {
			break
		}

		block = item.NewBlock(blk)
		if err := d.storeBlock(block); err != nil {
			return err
		}

		n += 1
	}
	log_.Debug("block digested to end", "height", block.Height, "count", n)

	return nil
}

func (d *Digest) storeBlock(block item.Block) (err error) {
	var txHashes []string
	txHashes = append(txHashes, block.Transactions...)
	txHashes = append(txHashes, block.ProposerTransaction)

	txs, err := sebak.GetTransactions(d.sst, txHashes...)
	if err != nil {
		log.Error("failed to get transactions from block", "block", block.Height, "error", err)
		return
	}

	st, err := d.st.OpenBatch()
	if err != nil {
		return err
	}

	if err = d.saveBlock(st, block, txs); err != nil {
		if err == sebakerrors.BlockAlreadyExists {
			log.Warn("block already exists", "block", block.Height)
		} else {
			log.Error("failed to save block and transactions", "block", block.Height, "error", err)
			st.Discard()
			return
		}
	}

	log.Debug("block saved", "block", block.Height, "txs", len(txHashes))
	return st.Commit()
}

func (d *Digest) logInsertedData() {
	var blocks, transactions, internals, accounts uint64

	{ // internals
		iterFunc, closeFunc := d.st.GetIterator(
			storage.InternalPrefix,
			sebakstorage.NewDefaultListOptions(false, nil, 0),
		)
		defer closeFunc()

		for {
			if _, hasNext := iterFunc(); !hasNext {
				break
			}
			internals += 1
		}

	}

	{ // blocks
		iterFunc, closeFunc := d.st.GetIterator(
			storage.BlockPrefix,
			sebakstorage.NewDefaultListOptions(false, nil, 0),
		)
		defer closeFunc()

		for {
			if _, hasNext := iterFunc(); !hasNext {
				break
			}
			blocks += 1
		}
	}

	{ // transactions
		iterFunc, closeFunc := d.st.GetIterator(
			storage.TransactionPrefix,
			sebakstorage.NewDefaultListOptions(false, nil, 0),
		)
		defer closeFunc()

		for {
			if _, hasNext := iterFunc(); !hasNext {
				break
			}
			transactions += 1
		}

	}

	{ // accounts
		iterFunc, closeFunc := d.st.GetIterator(
			storage.AccountPrefix,
			sebakstorage.NewDefaultListOptions(false, nil, 0),
		)
		defer closeFunc()

		for {
			if _, hasNext := iterFunc(); !hasNext {
				break
			}
			accounts += 1
		}

	}

	log.Debug(
		"data inserted",
		"blocks", blocks,
		"transactions", transactions,
		"accounts", accounts,
		"internals", internals,
	)
}

func (d *Digest) saveAccounts(st *storage.Storage, addresses ...string) error {
	var count int
	if len(addresses) < 1 {
		options := sebakstorage.NewDefaultListOptions(false, nil, 0)
		iterFunc, closeFunc := sebak.GetAccounts(d.sst, options)
		defer closeFunc()

		for {
			ac, hasNext := iterFunc()
			if !hasNext {
				break
			}
			// TODO remove
			log.Debug("> account", "ac", ac)
			if err := ac.Save(st); err != nil {
				return err
			}

			count += 1
		}
	} else {
		for _, address := range addresses {
			if address == d.genesisSource {
				continue
			}

			if ac, err := sebak.GetAccount(d.sst, address); err != nil {
				log.Error("failed to get account from sebak", "address", address)
				return err
			} else if err := ac.Save(st); err != nil {
				log.Error("failed to save account from sebak", "address", address)
				return err
			}
			count += 1
		}
	}

	log.Debug("save accounts", "count", count)

	return nil
}

func (d *Digest) saveBlock(st *storage.Storage, block item.Block, txs []item.Transaction) error {
	var addresses []string
	for _, tx := range txs {
		// TODO remove
		log.Debug("> tx", "tx", tx)
		if err := tx.Save(st); err != nil {
			return err
		}
		if d.initialize {
			addresses = append(addresses, tx.AllAccounts()...)
		}
	}

	// TODO remove
	log.Debug("> block", "block", block)
	if err := block.Save(st); err != nil {
		return err
	}

	if d.initialize {
		if err := d.saveAccounts(st, addresses...); err != nil {
			log.Error("failed to save accounts", "block", block.Height, "error", err, "txs", txs)
			return err
		}
	}

	return nil
}
