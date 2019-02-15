package cachebackend

import (
	"sync"
	"time"
)

type Item struct {
	Key    string
	Value  interface{}
	Expire time.Time
}

type SyncMap struct {
	m *sync.Map
}

func NewSyncMap() *SyncMap {
	return &SyncMap{m: &sync.Map{}}
}

func (s *SyncMap) Has(key string) (bool, error) {
	_, found := s.m.Load(key)
	return found, nil
}

func (s *SyncMap) Get(key string) (interface{}, error) {
	b, found := s.m.Load(key)
	if !found {
		return nil, CacheItemNotFound
	}
	item := b.(Item)
	if item.Expire.Before(time.Now()) {
		return nil, CacheItemNotFound
	}

	return item.Value, nil
}

func (s *SyncMap) Set(key string, v interface{}, expire time.Duration) error {
	s.m.Store(key, Item{Key: key, Value: v, Expire: time.Now().Add(expire)})
	return nil
}

func (s *SyncMap) Delete(key string) error {
	s.m.Delete(key)
	return nil
}
