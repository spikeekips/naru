package storage

import (
	sebakstorage "boscoin.io/sebak/lib/storage"
)

type Storage struct {
	*sebakstorage.LevelDBBackend
}

func NewStorage(nst *sebakstorage.LevelDBBackend) *Storage {
	return &Storage{
		LevelDBBackend: nst,
	}
}

func (s *Storage) OpenBatch() (*Storage, error) {
	raw, err := s.LevelDBBackend.OpenBatch()
	if err != nil {
		return nil, err
	}

	return NewStorage(raw), nil
}
