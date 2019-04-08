package storage

import (
	leveldbIterator "github.com/syndtr/goleveldb/leveldb/iterator"
)

type IterItem struct {
	N     uint64
	Key   []byte
	Value []byte
}

func NewIterItemFromIterator(iter leveldbIterator.Iterator) IterItem {
	ik := iter.Key()
	k := make([]byte, len(ik))
	copy(k, ik)

	iv := iter.Value()
	v := make([]byte, len(iv))
	copy(v, iv)

	return IterItem{
		Key:   k,
		Value: v,
	}
}

type SEBAKDBGetIteratorResult struct {
	Limit uint64     `json:"limit"`
	Items []IterItem `json:"items"`
}
