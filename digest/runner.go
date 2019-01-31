package digest

import (
	"sync"
	"time"

	sebakblock "boscoin.io/sebak/lib/block"
	sebakcommon "boscoin.io/sebak/lib/common"
	sebakkeypair "boscoin.io/sebak/lib/common/keypair"
	sebakerrors "boscoin.io/sebak/lib/errors"
	sebaknode "boscoin.io/sebak/lib/node"

	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
	"github.com/spikeekips/naru/storage/item"
)

var farBlockHeight uint64 = 1000

type BaseDigestRunner struct {
	sync.RWMutex
	st                  *storage.Storage
	sst                 *sebak.Storage
	sebakInfo           sebaknode.NodeInfo
	storedRemoteBlock   sebakblock.Block
	lastLocalBlock      item.Block
	TestLastRemoteBlock uint64
}

func (d *BaseDigestRunner) GenesisSource() string {
	kp := sebakkeypair.Master(string(d.sebakInfo.Policy.NetworkID))
	return kp.Address()
}

func (d *BaseDigestRunner) LastLocalBlock() item.Block {
	d.RLock()
	defer d.RUnlock()

	return d.lastLocalBlock
}

func (d *BaseDigestRunner) setLastLocalBlock(block item.Block) {
	d.Lock()
	defer d.Unlock()

	d.lastLocalBlock = block
}

func (d *BaseDigestRunner) StoredRemoteBlock() sebakblock.Block {
	d.RLock()
	defer d.RUnlock()

	return d.storedRemoteBlock
}

func (d *BaseDigestRunner) setStoredRemoteBlock(block sebakblock.Block) {
	d.Lock()
	defer d.Unlock()

	d.storedRemoteBlock = block
}

type InitializeDigestRunner struct {
	*BaseDigestRunner
	initialize bool
}

func NewInitializeDigestRunner(st *storage.Storage, sst *sebak.Storage, sebakInfo sebaknode.NodeInfo) *InitializeDigestRunner {
	return &InitializeDigestRunner{
		BaseDigestRunner: &BaseDigestRunner{
			st:        st,
			sst:       sst,
			sebakInfo: sebakInfo,
		},
	}
}

func (d *InitializeDigestRunner) Run() (err error) {
	log.Debug("start InitializeDigestRunner")

	sst := d.sst.New()
	if err = sst.Provider().Open(); err != nil {
		if err != sebak.ProviderNotClosedError {
			return
		}
	}

	defer func() {
		sst.Provider().Close()
	}()

	var lastRemoteBlock sebakblock.Block
	{ // get last remote block
		if d.TestLastRemoteBlock > 0 {
			lastRemoteBlock, err = sebak.GetBlockByHeight(sst, d.TestLastRemoteBlock)
		} else {
			lastRemoteBlock, err = sebak.GetLastBlock(sst)
		}
		if err != nil {
			return
		}
	}

	defer func() {
		if err != nil {
			return
		}
		d.setStoredRemoteBlock(lastRemoteBlock)
	}()

	{ // get last local block
		var block item.Block
		block, err = item.GetLastBlock(d.st)
		if err != nil {
			if e, ok := err.(*sebakerrors.Error); ok {
				if e.Code != sebakerrors.StorageRecordDoesNotExist.Code {
					return
				}
			}
		} else {
			d.setLastLocalBlock(block)
		}
	}

	if d.LastLocalBlock().Height < 1 {
		d.initialize = true
	}

	log.Debug(
		"last blocks",
		"local", d.LastLocalBlock().Height,
		"remote", lastRemoteBlock.Height,
		"initialize", d.initialize,
	)

	if d.LastLocalBlock().Height == lastRemoteBlock.Height {
		log.Debug(
			"local block reached to the remote block",
			"local", d.LastLocalBlock().Height,
			"remote", lastRemoteBlock.Height,
		)
		return nil
	}

	digest := NewDigest(d.st, d.sst, d.GenesisSource(), d.LastLocalBlock().Height, lastRemoteBlock.Height, d.initialize)
	if err = digest.Open(); err != nil {
		log.Error("failed to open digest", "error", err)
		return
	}
	defer digest.Close()

	if err = digest.Digest(); err != nil {
		return
	}

	d.setLastLocalBlock(item.NewBlock(lastRemoteBlock))

	return nil
}

type WatchDigestRunner struct {
	*BaseDigestRunner
	start uint64
}

func NewWatchDigestRunner(st *storage.Storage, sst *sebak.Storage, sebakInfo sebaknode.NodeInfo, start uint64) *WatchDigestRunner {
	return &WatchDigestRunner{
		BaseDigestRunner: &BaseDigestRunner{
			st:        st,
			sst:       sst,
			sebakInfo: sebakInfo,
		},
		start: start,
	}
}

// Run runs to watch and follow up the last remote block from sebak. By default,
// Run does not run if the local block is far behind from the remote
// block(`farBlockHeight`).
func (w *WatchDigestRunner) Run(force bool) error {
	sst := w.sst.New()
	if err := sst.Provider().Open(); err != nil {
		if err != sebak.ProviderNotClosedError {
			log.Error("failed to open sebak provider", "error", err)
			return err
		}
	}

	var lastRemoteBlock sebakblock.Block
	{ // get last remote block
		var err error
		lastRemoteBlock, err = sebak.GetLastBlock(sst)
		if err != nil {
			sst.Provider().Close()
			return err
		}
	}

	var startRemoteBlock sebakblock.Block
	if w.start > sebakcommon.GenesisBlockHeight { // get start block
		var err error
		startRemoteBlock, err = sebak.GetBlockByHeight(sst, w.start)
		if err != nil {
			if e, ok := err.(*sebakerrors.Error); ok {
				if e.Code != sebakerrors.StorageRecordDoesNotExist.Code {
					sst.Provider().Close()
					return err
				}
			}
			startRemoteBlock = lastRemoteBlock
		}
	}
	sst.Provider().Close()

	{ // get last local block
		block, err := item.GetLastBlock(w.st)
		if err != nil {
			return err
		}
		w.setLastLocalBlock(block)
	}

	log.Debug(
		"start WatchDigestRunner",
		"start", startRemoteBlock.Height,
		"remote", lastRemoteBlock.Height,
	)

	if startRemoteBlock.Height < lastRemoteBlock.Height { // follow up
		if !force && lastRemoteBlock.Height-startRemoteBlock.Height >= farBlockHeight {
			log.Error(
				"local block is too far from the remote block",
				"local", startRemoteBlock,
				"remote", lastRemoteBlock.Height,
			)
		}

		go w.followup(startRemoteBlock.Height, lastRemoteBlock.Height-1)
	}

	w.watchLatestBlocks()

	return nil
}

func (w *WatchDigestRunner) watchLatestBlocks() {
	log.Debug("start watchLatestBlocks")

	for {
		if err := w.watchLatestBlock(); err != nil {
			log.Error("something wrong watchLatestBlock", "error", err)
			time.Sleep(time.Second * 2)
			continue
		}

		time.Sleep(time.Millisecond * 300)
	}
}

func (w *WatchDigestRunner) followup(start, end uint64) error {
	log.Debug("start to follow up", "start", start, "end", end)

	digest := NewDigest(w.st, w.sst, w.GenesisSource(), start, end, false)
	if err := digest.Open(); err != nil {
		log.Error("failed to open digest", "error", err)
		return err
	}
	err := digest.Digest()
	digest.Close()
	if err != nil {
		return err
	}

	return nil
}

func (w *WatchDigestRunner) watchLatestBlock() error {
	sst := w.sst.New()
	if err := sst.Provider().Open(); err != nil {
		if err != sebak.ProviderNotClosedError {
			log.Error("failed to open sebak provider", "error", err)
			return err
		}
	}

	var err error
	var block sebakblock.Block
	block, err = sebak.GetLastBlock(sst)
	sst.Provider().Close()
	if err != nil {
		log.Error("failed to get last remote block", "error", err)
		return err
	}

	if block.Height == w.LastLocalBlock().Height {
		return nil
	}

	digest := NewDigest(w.st, w.sst, w.GenesisSource(), w.LastLocalBlock().Height, block.Height, false)
	if err := digest.Open(); err != nil {
		log.Error("failed to open digest", "error", err)
		return err
	}
	defer digest.Close()

	if err := digest.Digest(); err != nil {
		log.Error("failed to digest", "error", err)
		return err
	}

	w.setStoredRemoteBlock(block)
	w.setLastLocalBlock(item.NewBlock(block))

	return nil
}
