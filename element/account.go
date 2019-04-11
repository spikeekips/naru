package element

import (
	"fmt"

	sebakblock "boscoin.io/sebak/lib/block"
	sebakcommon "boscoin.io/sebak/lib/common"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/spikeekips/naru/storage"
)

type Account struct {
	Address      string             `json:"address"`
	Balance      sebakcommon.Amount `json:"balance"`
	SequenceID   uint64             `json:"sequence_id"`
	Linked       string             `json:"linked"`
	CodeHash     []byte             `json:"code_hash"`
	RootHash     string             `json:"root_hash"`
	CreatedBlock uint64             `json:"created_height"`
}

func NewAccount(ac sebakblock.BlockAccount) Account {
	return Account{
		Address:    ac.Address,
		Balance:    ac.Balance,
		SequenceID: ac.SequenceID,
		Linked:     ac.Linked,
		CodeHash:   ac.CodeHash,
		RootHash:   ac.RootHash.Hex(),
	}
}

func (a Account) Save(st storage.Storage) error {
	var f func(string, interface{}) error
	var created bool
	if found, err := st.Has(GetAccountKey(a.Address)); err != nil {
		return err
	} else if found {
		f = st.Update
	} else {
		f = st.Insert
		created = true
	}
	err := f(GetAccountKey(a.Address), a)
	if err == nil {
		st.Event("OnAfterSaveAccount", st, a, created)
		return nil
	}

	return err
}

func (a Account) BlockAccount() *sebakblock.BlockAccount {
	return &sebakblock.BlockAccount{
		Address:    a.Address,
		Balance:    a.Balance,
		SequenceID: a.SequenceID,
		Linked:     a.Linked,
		CodeHash:   a.CodeHash,
		RootHash:   ethcommon.HexToHash(a.RootHash),
	}
}

func GetAccountKey(address string) string {
	return fmt.Sprintf("%s%s", AccountPrefix, address)
}
