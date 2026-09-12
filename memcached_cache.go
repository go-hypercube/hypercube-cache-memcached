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

func (c *MemcachedCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	var value uint64
	var err error

	if delta >= 0 {
		value, err = c.client.Increment(key, uint64(delta))
	} else {
		value, err = c.client.Decrement(key, uint64(-delta))
	}

	if err == nil {
		return int64(value), nil
	}

	if !errors.Is(err, memcache.ErrCacheMiss) {
		return 0, err
	}

	initial := max(delta, 0)

	err = c.client. Add(&memcache.Item{
		Key:   key,
		Value: []byte(strconv.FormatInt(initial, 10)),
	})

	if err == nil {
		return initial, nil
	}

	if !errors.Is(err, memcache.ErrNotStored) {
		return 0, err
	}

	// Someone else initialized the key.
	if delta >= 0 {
		value, err = c.client.Increment(key, uint64(delta))
	} else {
		value, err = c.client.Decrement(key, uint64(-delta))
	}

	return int64(value), err
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
