package sebak

import (
	"github.com/spikeekips/naru/storage"
)

type Storage struct {
	provider StorageProvider
}

func NewStorage(p StorageProvider) *Storage {
	return &Storage{provider: p}
}

func (s *Storage) Provider() StorageProvider {
	return s.provider
}

func (s *Storage) New() *Storage {
	return NewStorage(s.provider.New())
}

func (s *Storage) Close() error {
	return s.provider.Close()
}

func (s *Storage) Has(key string) (bool, error) {
	return s.provider.Has(key)
}

func (s *Storage) GetRaw(key string) ([]byte, error) {
	return s.provider.Get(key)
}

func (s *Storage) Get(key string, i interface{}) ([]byte, error) {
	b, err := s.GetRaw(key)
	if err != nil {
		return nil, err
	}

	return b, storage.Deserialize(b, i)
}

func (s *Storage) GetIterator(prefix string, options storage.ListOptions) (func() (storage.IterItem, bool), func()) {
	return s.provider.GetIterator(prefix, options)
}
