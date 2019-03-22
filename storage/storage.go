package storage

import (
	"github.com/spikeekips/naru/query"
)

type Storage interface {
	Batch() BatchStorage
	Has(string) (bool, error)
	Get(string, interface{}) error
	Iterator(string, interface{}, ListOptions) (func() (Record, bool /* has next */) /* close func */, func())
	Insert(string, interface{}) error
	Update(string, interface{}) error
	Delete(string) error
	MultipleInsert(...Value) error
	MultipleUpdate(...Value) error
	MultipleDelete(...string) error
	Event(string, ...interface{})
}

type FilterableStorage interface {
	Storage
	Filter(query.Query, ListOptions) (func(interface{} /* result item */, bool /* has next */), func() /* close func */)
	FilterUpdate(query.Query, []byte) ( /* list of keys */ []string, error)
	FilterDelete(query.Query) ( /* list of keys */ []string, error)
}

type BatchStorage interface {
	Storage
	Write() error
	Cancel() error
}
