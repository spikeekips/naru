package resourcev1

import (
	"strings"

	sebakresource "boscoin.io/sebak/lib/node/runner/api/resource"
	"github.com/nvellon/hal"

	"github.com/spikeekips/naru/element"
)

type Block struct {
	b *element.Block
}

func NewBlock(b *element.Block) *Block {
	blk := &Block{
		b: b,
	}
	return blk
}

func (blk Block) GetMap() hal.Entry {
	b := blk.b
	return hal.Entry{
		"version":              b.Header.Version,
		"hash":                 b.Hash,
		"height":               b.Header.Height,
		"prev_block_hash":      b.Header.PrevBlockHash,
		"transactions_root":    b.Header.TransactionsRoot,
		"confirmed":            b.Confirmed,
		"proposer":             b.Proposer,
		"proposed_time":        b.Header.ProposedTime,
		"proposer_transaction": b.ProposerTransaction,
		"round":                b.Round,
		"transactions":         b.Transactions,
	}
}

func (blk Block) Resource() *hal.Resource {
	r := hal.NewResource(blk, blk.LinkSelf())
	return r
}

func (blk Block) LinkSelf() string {
	return strings.Replace(sebakresource.URLBlocks, "{id}", blk.b.Hash, -1)
}
