package geecache

import (
	"github.com/real3stone/geecache/geecache/lru"
	"sync"
)

//封装cache，使得可并发
type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) Add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) Get(key string) (value ByteView, ok bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok { //成功找到
		return v.(ByteView), ok
	}

	return
}
