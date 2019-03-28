package element

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

func (a Account) Save(st storage.Storage) error {
	var f func(string, interface{}) error
	var event string
	if found, err := st.Has(GetAccountKey(a.Address)); err != nil {
		return err
	} else if found {
		f = st.Update
		//event = EventUpdateAccount
	} else {
		f = st.Insert
		//event = EventNewAccount
	}
	err := f(GetAccountKey(a.Address), a)
	if err == nil {
		st.Event("OnAfterSaveAccount "+event, st, a)
		return nil
	}

	return err
}

func GetAccountKey(address string) string {
	return fmt.Sprintf("%s%s", AccountPrefix, address)
}
