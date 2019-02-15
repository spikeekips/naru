package cachebackend

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type GoCache struct {
	c *gocache.Cache
}

func NewGoCache() *GoCache {
	c := gocache.New(5*time.Minute, 10*time.Minute)

	return &GoCache{c: c}
}

func (s *GoCache) Has(key string) (bool, error) {
	_, found := s.c.Get(key)
	return found, nil
}

func (s *GoCache) Get(key string) (interface{}, error) {
	v, found := s.c.Get(key)
	if !found {
		return nil, CacheItemNotFound
	}

	return v, nil
}

func (s *GoCache) Set(key string, v interface{}, expire time.Duration) error {
	s.c.Set(key, v, expire)
	return nil
}

func (s *GoCache) Delete(key string) error {
	s.c.Delete(key)
	return nil
}
