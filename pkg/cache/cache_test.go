package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	t.Parallel()

	const (
		coof      = 100
		timeLimit = 10 * time.Second
		timeWrite = 10 * coof * time.Millisecond
		timeRead  = 5 * coof * time.Millisecond
		timeDel   = 20 * coof * time.Millisecond
	)

	c := NewCache[struct{}]()
	wg := sync.WaitGroup{}
	fns := make([]func(), 0)
	ctx, cancel := context.WithTimeout(t.Context(), timeLimit)

	t.Cleanup(cancel)
	t.Cleanup(func() {
		t.Logf("cache size before cleanup: %d", c.Size())
		c.Cleanup()
		t.Log("cache cleaned up")
		t.Logf("cache size after cleanup: %d", c.Size())
	})

	fns = append(fns, func() {
		defer wg.Done()

		ticker := time.NewTicker(timeWrite)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				key := time.Now().Format(time.DateTime)
				c.Add(key, struct{}{})
				t.Logf("add key: %s", key)
			}
		}
	}, func() {
		defer wg.Done()

		ticker := time.NewTicker(timeRead)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				key := time.Now().Format(time.DateTime)
				_, ok := c.Get(key)
				t.Logf("get key (%s): %v", key, ok)
			}
		}
	}, func() {
		defer wg.Done()

		ticker := time.NewTicker(timeDel)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				key := time.Now().Format(time.DateTime)
				c.Del(key)
				t.Logf("del key: %s", key)
			}
		}
	})

	wg.Add(len(fns))
	for _, fn := range fns {
		go fn()
	}

	wg.Wait()
}
