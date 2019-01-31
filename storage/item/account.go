package item

import (
	"fmt"

	sebakblock "boscoin.io/sebak/lib/block"

	"github.com/spikeekips/naru/storage"
)

type Account struct {
	sebakblock.BlockAccount
}

func NewAccount(ac sebakblock.BlockAccount) Account {
	return Account{BlockAccount: ac}
}

func (a Account) Save(st *storage.Storage) error {
	var f func(string, interface{}) error
	if found, err := st.Has(GetAccountKey(a.Address)); err != nil {
		return err
	} else if found {
		f = st.Set
	} else {
		f = st.New
	}
	return f(GetAccountKey(a.Address), a)
}

func GetAccountKey(address string) string {
	return fmt.Sprintf("%s%s", storage.AccountPrefix, address)
}

func GetAccount(st *storage.Storage, address string) (ac Account, err error) {
	err = st.Get(GetAccountKey(address), &ac)
	return
}
