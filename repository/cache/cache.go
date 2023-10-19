package cache

import (
	"strings"
	"sync"
	"time"

	"go.uber.org/config"
)

// Cache stores arbitrary data with expiration time.
type Cache struct {
	items sync.Map
	close chan struct{}
}

// An item represents arbitrary data with expiration time.
type item struct {
	data    interface{}
	expires int64
}

// New creates a new cache that asynchronously cleans
// expired entries after the given time passes.
func New(cfg config.Provider) *Cache {
	var durationCfg durationConfig
	cfg.Get("cache").Populate(&durationCfg)
	duration := durationMap(durationCfg)
	// cleaningInterval
	cache := &Cache{
		close: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now().UnixNano()

				cache.items.Range(func(key, value interface{}) bool {
					item := value.(item)

					if item.expires > 0 && now > item.expires {
						cache.items.Delete(key)
					}

					return true
				})

			case <-cache.close:
				return
			}
		}
	}()

	return cache
}

type durationConfig struct {
	interval int64
	duration string
}

func durationMap(p durationConfig) time.Duration {
	durationMap := map[string]time.Duration{
		"second": time.Second,
		"minute": time.Minute,
		"hour":   time.Hour,
	}
	if value, ok := durationMap[strings.ToLower(p.duration)]; ok {
		return time.Duration(p.interval) * value
	}
	return time.Duration(p.interval) * time.Minute
}

// Range calls f sequentially for each key and value present in the cache.
// If f returns false, range stops the iteration.
func (cache *Cache) Range(f func(key, value interface{}) bool) {
	now := time.Now().UnixNano()

	fn := func(key, value interface{}) bool {
		item := value.(item)

		if item.expires > 0 && now > item.expires {
			return true
		}

		return f(key, item.data)
	}

	cache.items.Range(fn)
}

// Get gets the value for the given key.
func (cache *Cache) Get(key interface{}) (interface{}, bool) {
	obj, exists := cache.items.Load(key)

	if !exists {
		return nil, false
	}

	item := obj.(item)

	if item.expires > 0 && time.Now().UnixNano() > item.expires {
		return nil, false
	}

	return item.data, true
}

// Set sets a value for the given key with an expiration duration.
// If the duration is 0 or less, it will be stored forever.
func (cache *Cache) Set(key interface{}, value interface{}, duration time.Duration) {
	var expires int64

	if duration > 0 {
		expires = time.Now().Add(duration).UnixNano()
	}

	cache.items.Store(key, item{
		data:    value,
		expires: expires,
	})
}

// Delete deletes the key and its value from the cache.
func (cache *Cache) Delete(key interface{}) {
	cache.items.Delete(key)
}

// Close closes the cache and frees up resources.
func (cache *Cache) Close() {
	cache.close <- struct{}{}
	cache.items = sync.Map{}
}
