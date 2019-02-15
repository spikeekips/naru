package storage

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"

	sebakstorage "boscoin.io/sebak/lib/storage"
)

type Storage struct {
	sync.RWMutex
	*sebakstorage.LevelDBBackend
}

func NewStorage(st *sebakstorage.LevelDBBackend) *Storage {
	if core, ok := st.Core.(*leveldb.DB); ok {
		st = &sebakstorage.LevelDBBackend{
			DB:   st.DB,
			Core: NewLevelDBCore(core),
		}
	}

	return &Storage{LevelDBBackend: st}
}

func (s *Storage) OpenBatch() (*Storage, error) {
	raw, err := s.LevelDBBackend.OpenBatch()
	if err != nil {
		return nil, err
	}

	raw.Core = NewBatchCore(raw.Core.(*sebakstorage.BatchCore))

	return NewStorage(raw), nil
}

func (s *Storage) Discard() error {
	if err := s.LevelDBBackend.Discard(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) AddEvent(event string, v ...interface{}) {
	core, ok := s.Core.(EventDBCore)
	if !ok {
		return
	}

	core.AddEvent(event, v...)
}

func (s *Storage) TriggerEvents() {
	core, ok := s.Core.(EventDBCore)
	if !ok {
		return
	}

	for event, v := range core.ClearEvents() {
		Observer.Trigger(event, v...)
	}
}
