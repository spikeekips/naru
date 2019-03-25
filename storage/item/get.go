package item

import (
	sebakstorage "boscoin.io/sebak/lib/storage"

	"github.com/spikeekips/naru/storage"
)

type Getter interface {
	Storage() storage.Storage
	Account( /* address */ string) (Account, error)
	Block( /* hash */ string) (Block, error)
	BlockByHeight( /* height */ uint64) (Block, error)
	LastBlock() (Block, error)
	BlocksIterator(
		iterfunc func() (sebakstorage.IterItem, bool),
		closefunc func(),
	) (
		func() (Block, bool, []byte),
		func(),
	)
	Operation( /* hash */ string) (op Operation, err error)
	OperationsByAccount( /* address */ string /* options */, storage.ListOptions) (
		func() (Operation, bool, []byte),
		func(),
	)
	ExistsTransaction( /* hash */ string) (bool, error)
	Transaction( /* hash */ string) (Transaction, error)
}
