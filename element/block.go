package element

import (
	"fmt"
	"time"

	sebakblock "boscoin.io/sebak/lib/block"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/storage"
)

type BlockHeader struct {
	Version          uint32
	PrevBlockHash    string
	TransactionsRoot string
	ProposedTime     time.Time
	Height           uint64
	TotalTxs         uint64
	TotalOps         uint64
}

type Block struct {
	Header              BlockHeader
	Transactions        []string
	ProposerTransaction string
	Hash                string
	Proposer            string
	Round               uint64
	Confirmed           time.Time
}

func NewBlock(block sebakblock.Block) Block {
	confirmed, _ := common.ParseISO8601(block.Confirmed)
	proposedTime, _ := common.ParseISO8601(block.Header.ProposedTime)

	return Block{
		Header: BlockHeader{
			Version:          block.Header.Version,
			PrevBlockHash:    block.Header.PrevBlockHash,
			TransactionsRoot: block.Header.TransactionsRoot,
			ProposedTime:     proposedTime,
			Height:           block.Header.Height,
			TotalTxs:         block.Header.TotalTxs,
			TotalOps:         block.Header.TotalOps,
		},
		Transactions:        block.Transactions,
		ProposerTransaction: block.ProposerTransaction,
		Hash:                block.Hash,
		Proposer:            block.Proposer,
		Round:               block.Round,
		Confirmed:           confirmed,
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
