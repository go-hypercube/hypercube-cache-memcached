# hypercube-cache-memcached

Memcached driver for [go-hypercube](https://github.com/go-hypercube/go-hypercube)'s `cache.Cache` interface, backed by [`bradfitz/gomemcache`](https://github.com/bradfitz/gomemcache).

## Install

```bash
go get go get github.com/go-hypercube/hypercube-cache-memcached
```

## Usage

```go
import (
	"github.com/bradfitz/gomemcache/memcache"
	memcachedcache "github.com/go-hypercube/hypercube-cache-memcached"
)

client := memcache.New("localhost:11211")
cache := memcachedcache.New(client)

app := hypercube.New(cfg, db, cache)
```

## Notes

- Implements the full `github.com/go-hypercube/go-hypercube/cache` interface: `Get`, `Set`, `Delete`, `Has`, `Increment`, `Expire`.
- Misses map to `cache.ErrNotFound`; negative TTLs return `cache.ErrInvalidTTL`.
- **`Increment` caveat:** Memcached only increments existing numeric keys — unlike Redis it does not create the key at 0. This driver emulates create-at-zero to match the interface contract, but the miss-then-set is **not atomic**: a concurrent `Increment` on the same missing key can race. Avoid relying on `Increment` for exact counters under high concurrency on cold keys with this driver.
- TTLs are converted to seconds (`int32`) per the Memcached protocol — sub-second TTLs are not supported and will round down.

## License

MIT
