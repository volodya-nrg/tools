package cache

import "sync"

type Cache[T any] struct {
	cache map[string]T
	mu    sync.RWMutex
}

func (c *Cache[T]) Add(key string, item T) {
	c.mu.Lock() // блокируем на запись, чтение
	defer c.mu.Unlock()

	c.cache[key] = item
}

func (c *Cache[T]) Get(key string) (T, bool) { //nolint:ireturn
	c.mu.RLock() // блокируем на запись, читать могут другие
	defer c.mu.RUnlock()

	item, ok := c.cache[key]
	return item, ok
}

func (c *Cache[T]) Del(key string) {
	c.mu.Lock() // блокируем на запись, чтение
	defer c.mu.Unlock()

	delete(c.cache, key)
}

func (c *Cache[T]) Cleanup() {
	c.mu.Lock() // блокируем на запись, чтение
	defer c.mu.Unlock()

	for k := range c.cache {
		delete(c.cache, k)
	}
}

func (c *Cache[T]) Size() int {
	c.mu.RLock() // блокируем на запись, читать могут другие
	defer c.mu.RUnlock()

	return len(c.cache)
}

func NewCache[T any]() *Cache[T] {
	return &Cache[T]{
		cache: make(map[string]T),
	}
}
