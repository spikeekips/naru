package leveldbstorage

import (
	"os"
	"reflect"

	"github.com/syndtr/goleveldb/leveldb"
	leveldbStorage "github.com/syndtr/goleveldb/leveldb/storage"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/storage"
)

type Storage struct {
	l    *leveldb.DB
	path string
}

func NewStorage(c *config.LevelDBStorage) (*Storage, error) {
	var db *leveldb.DB

	switch c.Scheme {
	case "memory":
		d, err := leveldb.Open(leveldbStorage.NewMemStorage(), nil)
		if err != nil {
			return nil, setError(err)
		}
		db = d
	case "file":
		d, err := leveldb.OpenFile(c.RealPath, nil)
		if err != nil {
			return nil, setError(err)
		}
		db = d
	}

	return &Storage{l: db, path: c.RealPath}, nil
}

func (b *Storage) Core() *leveldb.DB {
	return b.l
}

func (b *Storage) Close() error {
	return b.l.Close()
}

func (b *Storage) Initialize() error {
	if len(b.path) < 1 {
		return nil
	}

	return os.RemoveAll(b.path)
}

func (b *Storage) Batch() (storage.BatchStorage, error) {
	return NewBatch(b)
}

func (b *Storage) Write() error {
	return storage.NotBatchStorage.New()
}

func (b *Storage) Has(k string) (bool, error) {
	ok, err := b.l.Has(makeKey(k), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, setError(err)
	}

	return ok, nil
}

func (b *Storage) MustExist(k string) error {
	exists, err := b.Has(k)
	if err != nil {
		return err
	} else if !exists {
		return storage.NotFound.New()
	}

	return nil
}

func (b *Storage) MustNotExist(k string) error {
	exists, err := b.Has(k)
	if err != nil {
		return err
	} else if exists {
		return storage.AlreadyExists.New()
	}

	return nil
}

func (b *Storage) GetRaw(k string) ([]byte, error) {
	if err := b.MustExist(k); err != nil {
		return nil, err
	}

	return b.l.Get(makeKey(k), nil)
}

func (b *Storage) Get(k string, v interface{}) error {
	o, err := b.GetRaw(k)
	if err != nil {
		return setError(err)
	}

	return storage.Deserialize(o, v)
}

func (b *Storage) IteratorRaw(prefix string, options storage.ListOptions) (func() (storage.IterItem, bool, error), func(), error) {
	var reverse = false
	var cursor []byte
	var limit uint64 = 0
	if options != nil {
		reverse = options.Reverse()
		cursor = options.Cursor()
		limit = options.Limit()
	}

	var dbRange *leveldbUtil.Range
	if len(prefix) > 0 {
		dbRange = leveldbUtil.BytesPrefix(makeKey(prefix))
	}

	iter := b.l.NewIterator(dbRange, nil)

	var funcNext func() bool
	var seek func() bool

	if reverse {
		funcNext = iter.Prev
		if cursor == nil {
			seek = iter.Last
		} else {
			seek = func() bool {
				iter.Seek(cursor)
				return funcNext()
			}
		}
	} else {
		funcNext = iter.Next
		if cursor == nil {
			seek = iter.First
		} else {
			seek = func() bool {
				iter.Seek(cursor)
				return funcNext()
			}
		}
	}

	var n uint64 = 0
	return func() (storage.IterItem, bool, error) {
			var next bool
			if n == 0 {
				next = seek()
			} else {
				next = funcNext()
			}

			n++

			if limit != 0 && n > limit {
				iter.Release()
				return storage.IterItem{}, false, nil
			}

			return storage.NewIterItemFromIterator(iter), next, nil
		},
		func() {
			iter.Release()
		}, nil
}

func (b *Storage) Iterator(prefix string, v interface{}, options storage.ListOptions) (func() (storage.Record, bool, error), func(), error) {
	iterFunc, closeFunc, err := b.IteratorRaw(prefix, options)
	if err != nil {
		return nil, nil, err
	}

	return func() (storage.Record, bool, error) {
			item, next, err := iterFunc()
			if err != nil {
				return storage.Record{}, false, err
			} else if !next {
				return storage.Record{}, false, nil
			}

			nv := reflect.New(reflect.TypeOf(v)).Interface()
			if err := storage.Deserialize(item.Value, nv); err != nil {
				return storage.Record{}, false, err
			}

			return storage.NewRecord(
				string(item.Key),
				reflect.ValueOf(nv).Elem().Interface(),
			), next, nil
		},
		closeFunc,
		nil
}

func (b *Storage) Insert(k string, v interface{}) error {
	if err := b.MustNotExist(k); err != nil {
		return err
	}

	encoded, err := storage.Serialize(v)
	if err != nil {
		return setError(err)
	}

	return setError(b.l.Put(makeKey(k), encoded, nil))
}

func (b *Storage) Update(k string, v interface{}) error {
	if err := b.MustExist(k); err != nil {
		return err
	}

	encoded, err := storage.Serialize(v)
	if err != nil {
		return setError(err)
	}

	return setError(b.l.Put(makeKey(k), encoded, nil))
}

func (b *Storage) Delete(k string) error {
	if err := b.MustExist(k); err != nil {
		return err
	}

	return setError(b.l.Delete(makeKey(k), nil))
}

func (b *Storage) Event(event string, values ...interface{}) {
	storage.Observer.Trigger(event, values...)
}

func setError(err error) error {
	if err == nil {
		return nil
	}

	return LevelDBCoreError.NewFromError(err)
}

func makeKey(k string) []byte {
	return []byte(k)
}
