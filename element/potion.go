package element

import (
	"github.com/spikeekips/naru/storage"
)

type Potion interface {
	Storage() storage.Storage
	Account( /* address */ string) (Account, error)
	Accounts(string, storage.ListOptions) (
		func() (Account, bool, []byte),
		func(),
	)
	Block( /* hash */ string) (Block, error)
	BlockByHeight( /* height */ uint64) (Block, error)
	LastBlock() (Block, error)
	BlocksByHeight(start, end uint64) (
		func() (Block, bool, []byte),
		func(),
	)
	Operation( /* hash */ string) (op Operation, err error)
	OperationsByAccount( /* address */ string /* options */, storage.ListOptions) (
		func() (Operation, bool, []byte),
		func(),
	)
	OperationsByTransaction( /* hash */ string /* options */, storage.ListOptions) (
		func() (Operation, bool, []byte),
		func(),
	)
	ExistsTransaction( /* hash */ string) (bool, error)
	Transaction( /* hash */ string) (Transaction, error)
	TransactionsByBlock( /* block hash */ string, storage.ListOptions) (
		func() (Transaction, bool, []byte),
		func(),
	)
	TransactionsByAccount( /* address */ string, storage.ListOptions) (
		func() (Transaction, bool, []byte),
		func(),
	)
	OperationsByHeight(start, end uint64) (
		func() (Operation, bool, []byte),
		func(),
	)
	BlockStat() (BlockStat, error)
}
