package leveldbstorage

import (
	"strings"
	"sync"

	"boscoin.io/sebak/lib/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/storage"
)

type Batch struct {
	sync.RWMutex
	s      *Storage
	b      *leveldb.Batch
	events []common.EventItem
}

func NewBatch(s *Storage) *Batch {
	return &Batch{
		s: s,
		b: new(leveldb.Batch),
	}
}

func (b *Batch) Initialize() error {
	return nil
}

func (b *Batch) Has(k string) (bool, error) {
	return b.s.Has(k)
}

func (b *Batch) Get(k string, v interface{}) error {
	return b.s.Get(k, v)
}

func (b *Batch) Iterator(prefix string, v interface{}, options storage.ListOptions) (func() (storage.Record, bool), func()) {
	return b.s.Iterator(prefix, v, options)
}

func (b *Batch) Insert(k string, v interface{}) error {
	if err := b.s.MustNotExist(k); err != nil {
		return err
	}

	encoded, err := storage.Serialize(v)
	if err != nil {
		return setError(err)
	}

	b.b.Put(makeKey(k), encoded)
	return nil
}

func (b *Batch) Update(k string, v interface{}) error {
	if err := b.s.MustExist(k); err != nil {
		return err
	}

	encoded, err := storage.Serialize(v)
	if err != nil {
		return setError(err)
	}

	b.b.Put(makeKey(k), encoded)
	return nil
}

func (b *Batch) Delete(k string) error {
	if err := b.s.MustExist(k); err != nil {
		return err
	}

	b.b.Delete(makeKey(k))
	return nil
}

func (b *Batch) MultipleInsert(items ...storage.Value) error {
	if len(items) < 1 {
		return setError(errors.New("empty values"))
	}

	for _, i := range items {
		if err := b.s.MustNotExist(i.Key); err != nil {
			return err
		}
	}

	var l [][2][]byte
	for _, i := range items {
		encoded, err := storage.Serialize(i.Value)
		if err != nil {
			return setError(err)
		}

		l = append(l, [2][]byte{makeKey(i.Key), encoded})
	}

	for _, i := range l {
		b.b.Put(i[0], i[1])
	}

	return nil
}

func (b *Batch) MultipleUpdate(items ...storage.Value) error {
	if len(items) < 1 {
		return setError(errors.New("empty values"))
	}

	for _, i := range items {
		if err := b.s.MustExist(i.Key); err != nil {
			return err
		}
	}

	var l [][2][]byte
	for _, i := range items {
		encoded, err := storage.Serialize(i.Value)
		if err != nil {
			return setError(err)
		}

		l = append(l, [2][]byte{makeKey(i.Key), encoded})
	}

	for _, i := range l {
		b.b.Put(i[0], i[1])
	}

	return nil
}

func (b *Batch) MultipleDelete(keys ...string) error {
	if len(keys) < 1 {
		return setError(errors.New("empty values"))
	}

	for _, i := range keys {
		if err := b.s.MustExist(i); err != nil {
			return err
		}
	}

	for _, i := range keys {
		b.b.Delete(makeKey(i))
	}

	return nil
}

func (b *Batch) Write() error {
	var events []common.EventItem
	{
		b.RLock()
		events = make([]common.EventItem, len(b.events))
		copy(events, b.events)
		b.RUnlock()
	}

	for _, e := range events { // only OnAfterSave event will be triggered
		var es []string
		for _, n := range strings.Fields(e.Events) {
			if !strings.HasPrefix(n, "OnAfterSave") {
				continue
			}
			es = append(es, n)
		}

		storage.Observer.Trigger(strings.Join(es, " "), e.Items...)
	}

	err := setError(b.s.Core().Write(b.b, nil))
	if err != nil {
		return err
	}

	for _, e := range events { // only non-OnAfterSave event will be triggered
		var events []string
		for _, n := range strings.Fields(e.Events) {
			if strings.HasPrefix(n, "OnAfterSave") {
				continue
			}
			events = append(events, n)
		}

		storage.Observer.Trigger(strings.Join(events, " "), e.Items...)
	}

	b.clearEvents()
	return nil
}

func (b *Batch) Cancel() error {
	defer b.clearEvents()

	b.b = new(leveldb.Batch)
	return nil
}

func (b *Batch) Close() error {
	b.Cancel()
	return b.s.Close()
}

func (b *Batch) Batch() storage.BatchStorage {
	return b
}

func (b *Batch) Event(event string, values ...interface{}) {
	b.Lock()
	defer b.Unlock()

	b.events = append(b.events, common.NewEventItem(event, values...))
}

func (b *Batch) clearEvents() {
	b.Lock()
	defer b.Unlock()

	b.events = nil
}
