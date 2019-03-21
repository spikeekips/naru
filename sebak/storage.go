package sebak

import (
	sebakstorage "boscoin.io/sebak/lib/storage"

	storage "github.com/spikeekips/naru/newstorage"
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

func (s *Storage) GetIterator(prefix string, options sebakstorage.ListOptions) (func() (sebakstorage.IterItem, bool), func()) {
	var cursor []byte = options.Cursor()
	var limit uint64
	var items []sebakstorage.IterItem
	var err error
	var closed bool

	var all int
	var n int
	var iterFunc func() (sebakstorage.IterItem, bool)
	iterFunc = func() (sebakstorage.IterItem, bool) {
		if closed {
			return sebakstorage.IterItem{}, false
		}

		if options.Limit() > 0 && uint64(all) >= options.Limit() {
			closed = true
			return sebakstorage.IterItem{}, false
		}

		if items == nil {
			options.SetCursor(cursor)
			limit, items, err = s.provider.GetIterator(prefix, options)
			n = 0
		}

		if err != nil {
			log.Error("failed GetIterator", "error", err)
			return sebakstorage.IterItem{}, false
		}
		if len(items) == 0 {
			return sebakstorage.IterItem{}, false
		}
		if len(items) >= n+1 {
			defer func() {
				n += 1
				all += 1
			}()
			return items[n], true
		}
		if int(limit) > len(items) {
			return sebakstorage.IterItem{}, false
		}

		items = nil

		return iterFunc()
	}

	return func() (sebakstorage.IterItem, bool) {
			item, hasNext := iterFunc()
			if !hasNext {
				return item, hasNext
			}
			cursor = item.Key
			return item, hasNext
		}, func() {
			closed = true
		}
}
