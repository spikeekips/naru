package leveldbstorage

import (
	"reflect"

	"boscoin.io/sebak/lib/errors"
	"github.com/syndtr/goleveldb/leveldb"
	leveldbStorage "github.com/syndtr/goleveldb/leveldb/storage"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/storage"
)

type Storage struct {
	l *leveldb.DB
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

	return &Storage{l: db}, nil
}

func (b *Storage) Core() *leveldb.DB {
	return b.l
}

func (b *Storage) Close() error {
	return b.l.Close()
}

func (b *Storage) Batch() storage.BatchStorage {
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

func (b *Storage) Get(k string, v interface{}) error {
	if err := b.MustExist(k); err != nil {
		return err
	}

	o, err := b.l.Get(makeKey(k), nil)
	if err != nil {
		return setError(err)
	}

	return storage.Deserialize(o, v)
}

func (b *Storage) Iterator(prefix string, v interface{}, options storage.ListOptions) (func() (storage.Record, bool), func()) {
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
	return func() (storage.Record, bool) {
			var exists bool
			if n == 0 {
				exists = seek()
			} else {
				exists = funcNext()
			}

			nv := reflect.New(reflect.TypeOf(v)).Interface()
			if err := storage.Deserialize(iter.Value(), nv); err != nil {
				return storage.Record{}, false
			}

			n++

			if limit != 0 && n > limit {
				exists = false
			}

			return storage.NewRecord(
				string(iter.Key()),
				reflect.ValueOf(nv).Elem().Interface(),
			), exists
		},
		func() {
			iter.Release()
		}
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

func (b *Storage) MultipleInsert(items ...storage.Value) error {
	if len(items) < 1 {
		return setError(errors.New("empty values"))
	}

	for _, i := range items {
		if err := b.MustNotExist(i.Key); err != nil {
			return err
		}
	}

	batch := new(leveldb.Batch)
	for _, i := range items {
		encoded, err := storage.Serialize(i.Value)
		if err != nil {
			return setError(err)
		}

		batch.Put(makeKey(i.Key), encoded)
	}

	return setError(b.l.Write(batch, nil))
}

func (b *Storage) MultipleUpdate(items ...storage.Value) error {
	if len(items) < 1 {
		return setError(errors.New("empty values"))
	}

	for _, i := range items {
		if err := b.MustExist(i.Key); err != nil {
			return err
		}
	}

	batch := new(leveldb.Batch)
	for _, i := range items {
		encoded, err := storage.Serialize(i.Value)
		if err != nil {
			return setError(err)
		}

		batch.Put(makeKey(i.Key), encoded)
	}

	return setError(b.l.Write(batch, nil))
}

func (b *Storage) MultipleDelete(keys ...string) error {
	if len(keys) < 1 {
		return setError(errors.New("empty values"))
	}

	for _, i := range keys {
		if err := b.MustExist(i); err != nil {
			return err
		}
	}

	batch := new(leveldb.Batch)
	for _, i := range keys {
		batch.Delete(makeKey(i))
	}

	return setError(b.l.Write(batch, nil))
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
