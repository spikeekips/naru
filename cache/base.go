package cache

import (
	"fmt"
	"time"

	cachebackend "github.com/spikeekips/naru/cache/backend"
)

type Cache struct {
	prefix  string
	backend cachebackend.Backend
}

func NewCache(prefix string, backend cachebackend.Backend) *Cache {
	return &Cache{prefix: prefix, backend: backend}
}

func (c *Cache) key(key string) string {
	return fmt.Sprintf("%s-%s", c.prefix, key)
}

func (c *Cache) Has(key string) (bool, error) {
	return c.backend.Has(c.key(key))
}

func (c *Cache) Get(key string) (interface{}, error) {
	return c.backend.Get(c.key(key))
}

func (c *Cache) Set(key string, v interface{}, expire time.Duration) error {
	if expire < 1 {
		expire = 0 // no expiration
	}

	return c.backend.Set(c.key(key), v, expire)
}

func (c *Cache) Delete(key string) error {
	return c.backend.Delete(c.key(key))
}
