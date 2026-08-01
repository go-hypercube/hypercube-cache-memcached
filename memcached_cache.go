package memcachedcache

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-hypercube/go-hypercube/cache"
)

type MemcachedCache struct {
	client *memcache.Client
}

func New(client *memcache.Client) *MemcachedCache {
	return &MemcachedCache{client: client}
}

func (c *MemcachedCache) Get(ctx context.Context, key string) (string, error) {
	item, err := c.client.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return "", cache.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return string(item.Value), nil
}

func (c *MemcachedCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if ttl < 0 {
		return cache.ErrInvalidTTL
	}
	return c.client.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: int32(ttl.Seconds()),
	})
}

func (c *MemcachedCache) Delete(ctx context.Context, key string) error {
	err := c.client.Delete(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return nil
	}
	return err
}

func (c *MemcachedCache) Has(ctx context.Context, key string) (bool, error) {
	_, err := c.client.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return false, nil
	}
	return err == nil, err
}

// Increment: Memcached only increments existing numeric keys — unlike
// Redis' INCRBY it does not create the key at zero. This emulates Redis'
// semantics, but the miss-then-set is NOT atomic (a concurrent Increment
// on the same missing key can race). See caveats below.
func (c *MemcachedCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	var newValue uint64
	var err error
	if delta >= 0 {
		newValue, err = c.client.Increment(key, uint64(delta))
	} else {
		newValue, err = c.client.Decrement(key, uint64(-delta))
	}
	if errors.Is(err, memcache.ErrCacheMiss) {
		initial := delta
		if initial < 0 {
			initial = 0
		}
		if setErr := c.client.Set(&memcache.Item{Key: key, Value: []byte(strconv.FormatInt(initial, 10))}); setErr != nil {
			return 0, setErr
		}
		return initial, nil
	}
	return int64(newValue), err
}


func (c *MemcachedCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if ttl < 0 {
		return cache.ErrInvalidTTL
	}
	err := c.client.Touch(key, int32(ttl.Seconds()))
	if errors.Is(err, memcache.ErrCacheMiss) {
		return cache.ErrNotFound
	}
	return err
}