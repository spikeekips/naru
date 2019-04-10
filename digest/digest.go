package digest

import (
	"fmt"
	"time"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebakerrors "boscoin.io/sebak/lib/errors"
	logging "github.com/inconshreveable/log15"

	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
)

type Digest struct {
	sst           *sebak.Storage
	potion        element.Potion
	genesisSource string
	blocksLimit   uint64
	start         uint64
	end           uint64
	initialize    bool
	maxWorkers    int
}

func NewDigest(sst *sebak.Storage, potion element.Potion, genesisSource string, start, end uint64, initialize bool, maxWorkers int, blocksLimit uint64) (*Digest, error) {
	if start > end {
		return nil, fmt.Errorf("invalid start and end range: %d - %d", start, end)
	}

	return &Digest{
		sst:           sst.New(),
		potion:        potion,
		genesisSource: genesisSource,
		start:         start,
		end:           end,
		initialize:    initialize,
		maxWorkers:    maxWorkers,
		blocksLimit:   blocksLimit,
	}, nil
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
	if d.end == d.start {
		return nil
	}

	numberOfWorkers := int((d.end - d.start) / d.blocksLimit)
	if numberOfWorkers < 1 {
		numberOfWorkers = 1
	} else if numberOfWorkers > d.maxWorkers {
		numberOfWorkers = d.maxWorkers
	}

	started := time.Now()
	log_ := log.New(logging.Ctx{
		"start":           d.start,
		"end":             d.end,
		"blocksLimit":     d.blocksLimit,
		"numberOfWorkers": numberOfWorkers,
		"maxWorkers":      d.maxWorkers,
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

		cursorSet = append(cursorSet, [2]uint64{start, end})
		start += d.blocksLimit
		countCursors += 1
	}

	go func() {
		for _, cursors := range cursorSet {
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
		batch, err := d.newBatch()
		if err != nil {
			return err
		}

		if err := d.saveAccounts(batch); err != nil {
			return err
		}
		if err := batch.Write(); err != nil {
			return err
		}
	}

	// TODO remove
	//d.logInsertedData()

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
	started := time.Now()
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

	options := storage.NewDefaultListOptions(
		false,
		cursor,
		d.blocksLimit,
	)
	iterFunc, closeFunc := sebak.GetBlocks(d.sst, options)
	defer closeFunc()

	var blocks []element.Block
	for {
		blk, hasNext := iterFunc()
		if !hasNext {
			break
		}
		if blk.Height > end {
			break
		}

		blocks = append(blocks, element.NewBlock(blk))
	}
	log_.Debug(
		"block fetched",
		"blocks", len(blocks),
		"fetch-elapsed", time.Now().Sub(started),
		"elapsed", time.Now().Sub(started),
	)

	digestStarted := time.Now()
	batch, err := d.newBatch()
	if err != nil {
		return err
	}
	defer batch.Close()

	var block element.Block
	for _, block = range blocks {
		if err := d.digestBlock(batch, block); err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 300)
	}

	log_.Debug(
		"block digested",
		"last-block", block.Height,
		"blocks", len(blocks),
		"elapsed", time.Now().Sub(started),
		"digest-elapsed", time.Now().Sub(digestStarted),
	)

	writeStarted := time.Now()
	if err := batch.Write(); err != nil {
		return err
	}

	log_.Debug(
		"blocks saved",
		"last-block", block.Height,
		"blocks", len(blocks),
		"elapsed", time.Now().Sub(started),
		"write-elapsed", time.Now().Sub(writeStarted),
	)

	return nil
}

func (d *Digest) digestBlock(st storage.Storage, block element.Block) error {
	var txHashes []string
	txHashes = append(txHashes, block.Transactions...)
	txHashes = append(txHashes, block.ProposerTransaction)

	txs, err := sebak.GetTransactions(d.sst, txHashes...)
	if err != nil {
		log.Error("failed to get transactions from block", "block", block.Height, "error", err)
		return err
	}

	err = d.saveBlock(st, block, txs)
	if err != nil {
		if err == sebakerrors.BlockAlreadyExists {
			log.Warn("block already exists", "block", block.Height)
			return nil
		} else {
			log.Error("failed to save block and transactions", "block", block.Height, "error", err)
			return err
		}
	}

	return nil
}

func (d *Digest) logInsertedData() {
	var blocks, transactions, internals, accounts uint64

	{ // internals
		iterFunc, closeFunc := d.potion.Storage().Iterator(
			element.InternalPrefix,
			"",
			storage.NewDefaultListOptions(false, nil, 0),
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
		iterFunc, closeFunc := d.potion.Storage().Iterator(
			element.BlockPrefix,
			element.Block{},
			storage.NewDefaultListOptions(false, nil, 0),
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
		iterFunc, closeFunc := d.potion.Storage().Iterator(
			element.TransactionPrefix,
			element.Transaction{},
			storage.NewDefaultListOptions(false, nil, 0),
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
		iterFunc, closeFunc := d.potion.Storage().Iterator(
			element.AccountPrefix,
			element.Account{},
			storage.NewDefaultListOptions(false, nil, 0),
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

func (d *Digest) saveAccounts(st storage.Storage, addresses ...string) error {
	var count int
	if len(addresses) < 1 {
		options := storage.NewDefaultListOptions(false, nil, 0)
		iterFunc, closeFunc := sebak.GetAccounts(d.sst, options)
		defer closeFunc()

		for {
			ac, hasNext := iterFunc()
			if !hasNext {
				break
			}
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

	log.Debug("accounts saved", "count", count)

	return nil
}

func (d *Digest) saveBlock(st storage.Storage, block element.Block, txs []element.TransactionMessage) error {
	var addresses []string
	for _, txm := range txs {
		tx := element.NewTransaction(txm.Transaction, block, txm.Raw)
		if err := tx.Save(st); err != nil {
			log.Error("failed to save transaction", "tx", tx.Hash, "error", err)
			return err
		}
		if !d.initialize {
			addresses = append(addresses, tx.AllAccounts()...)
		}
	}

	if err := block.Save(st); err != nil {
		return err
	}

	if !d.initialize {
		if err := d.saveAccounts(st, addresses...); err != nil {
			log.Error("failed to save accounts", "block", block.Height, "error", err, "txs", txs)
			return err
		}
	}

	return nil
}

func (d *Digest) newBatch() (storage.BatchStorage, error) {
	var batch storage.BatchStorage
	var err error

	for i := 0; i < 10; i++ {
		if batch, err = d.potion.Storage().Batch(); err == nil {
			return batch, nil
		}
		log.Warn("failed to create batch", "error", err, "t", i)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return nil, err
}
