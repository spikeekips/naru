package element

import (
	"fmt"

	sebakblock "boscoin.io/sebak/lib/block"

	"github.com/spikeekips/naru/storage"
)

type Block struct {
	sebakblock.Header
	Transactions        []string
	ProposerTransaction string
	Hash                string
	Proposer            string
	Round               uint64
	Confirmed           string
}

func NewBlock(block sebakblock.Block) Block {
	return Block{
		Header:              block.Header,
		Transactions:        block.Transactions,
		ProposerTransaction: block.ProposerTransaction,
		Hash:                block.Hash,
		Proposer:            block.Proposer,
		Round:               block.Round,
		Confirmed:           block.Confirmed,
	}
}

func (b Block) Save(st storage.Storage) error {
	if err := st.Insert(GetBlockKey(b.Hash), b); err != nil {
		return err
	}

	st.Event("OnAfterSaveBlock", st, b)

	return nil
}

func GetBlockKey(hash string) string {
	return fmt.Sprintf("%s%s", BlockPrefix, hash)
}
