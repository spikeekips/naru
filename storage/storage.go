package storage

type Storage interface {
	Initialize() error
	Batch() (BatchStorage, error)
	Has(string) (bool, error)
	Get(string, interface{}) error
	Iterator(string, interface{}, ListOptions) (func() (Record, bool /* has next */) /* close func */, func())
	Insert(string, interface{}) error
	Update(string, interface{}) error
	Delete(string) error
	Event(string, ...interface{})
}

type BatchStorage interface {
	Storage
	Write() error
	Cancel() error
	Close() error
}
