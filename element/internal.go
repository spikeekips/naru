package element

import (
	"fmt"

	"github.com/spikeekips/naru/storage"
)

type BlockStat struct {
	TotalSupply uint64
}

func NewBlockStat() BlockStat {
	return BlockStat{}
}

func (a BlockStat) Save(st storage.Storage) error {
	var f func(string, interface{}) error
	if found, err := st.Has(GetBlockStatKey()); err != nil {
		return err
	} else if found {
		f = st.Update
	} else {
		f = st.Insert
	}

	return f(GetBlockStatKey(), a)
}

func GetBlockStatKey() string {
	return fmt.Sprintf("%s-blockstat", InternalPrefix)
}
