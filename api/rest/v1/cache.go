package restv1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spikeekips/naru/cache"
	cachebackend "github.com/spikeekips/naru/cache/backend"
)

const (
	requestCacheKeyUseCache string = "cache-use"
	requestCacheKeyExpire   string = "cache-expire"
)

type CacheWriter struct {
	http.ResponseWriter
	buf    []byte
	status int
}

func NewCacheWriter(w http.ResponseWriter) *CacheWriter {
	return &CacheWriter{ResponseWriter: w}
}

func (w *CacheWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *CacheWriter) Write(p []byte) (n int, err error) {
	w.buf = append(w.buf, p...)
	return w.ResponseWriter.Write(p)
}

func (w *CacheWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}

	return w.status
}

func (w *CacheWriter) Buffer() []byte {
	return w.buf
}

type CacheHandlerCondFunc func(*CacheWriter, *http.Request) (time.Duration, bool)

type CacheItem struct {
	status int
	header http.Header
	body   []byte
}

func (c CacheItem) Write(w http.ResponseWriter) {
	for k, v := range c.header {
		for _, i := range v {
			w.Header().Add(k, i)
		}
	}
	w.Header().Set("X-SEBAK-CACHE", "1")

	w.WriteHeader(c.status)
	w.Write(c.body)
}

type CacheHandler struct {
	cch          *cache.Cache
	expire       time.Duration
	handler      func(http.ResponseWriter, *http.Request)
	cacheKeyFunc func(*http.Request) string
	postConds    []CacheHandlerCondFunc
}

func NewCacheHandler(cch *cache.Cache, expire time.Duration, handler func(http.ResponseWriter, *http.Request)) *CacheHandler {
	return &CacheHandler{
		cch:     cch,
		expire:  expire,
		handler: handler,
	}
}

func (c *CacheHandler) defaultCacheKey(r *http.Request) string {
	return fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)
}

func (c *CacheHandler) cacheKey(r *http.Request) string {
	return c.cacheKeyFunc(r)
}

func (c *CacheHandler) SetCacheKey(fn func(*http.Request) string) *CacheHandler {
	c.cacheKeyFunc = fn
	return c
}

func (c *CacheHandler) Handler() func(http.ResponseWriter, *http.Request) {
	if c.cacheKeyFunc == nil {
		c.cacheKeyFunc = c.defaultCacheKey
	}

	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := c.cacheKey(r)
		if len(cacheKey) < 1 {
			c.handler(w, r)
			return
		}

		if raw, err := c.cch.Get(cacheKey); err != nil {
			if err != cachebackend.CacheItemNotFound {
				jw := NewJSONWriter(w)
				jw.WriteObject(err)
				return
			}
		} else if item, ok := raw.(CacheItem); !ok {
			jw := NewJSONWriter(w)
			jw.WriteObject(fmt.Errorf("something wrong in cache"))
			return
		} else {
			item.Write(w)
			return
		}

		cw := NewCacheWriter(w)

		c.handler(cw, r)

		var expire time.Duration = c.expire
		var matched bool
		for _, cond := range c.postConds {
			if expire, matched = cond(cw, r); matched {
				break
			}
		}

		if !matched {
			return
		}

		if expire < 0 { // negative expire value means no-cache
			return
		}

		item := CacheItem{
			status: cw.Status(),
			header: cw.Header(),
			body:   cw.Buffer(),
		}

		c.cch.Set(cacheKey, item, expire)
	}
}

func (c *CacheHandler) Status(expire time.Duration, status ...int) *CacheHandler {
	c.postConds = append(
		c.postConds,
		func(cw *CacheWriter, r *http.Request) (time.Duration, bool) {
			if len(status) < 1 {
				return expire, true
			}

			for _, s := range status {
				if cw.Status() == s {
					return expire, true
				}
			}

			return 0, false
		},
	)

	return c
}
