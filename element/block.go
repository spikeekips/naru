package element

import (
	"fmt"

	sebakblock "boscoin.io/sebak/lib/block"

	"github.com/spikeekips/naru/storage"
)

type Block struct {
	sebakblock.Block
}

func NewBlock(block sebakblock.Block) Block {
	return Block{Block: block}
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
