package cacheutil

import (
	"github.com/acernus18/dwarf/serializeutil"
	"github.com/bluele/gcache"
	"time"
)

func Load[T any](cache gcache.Cache, key string, expire time.Duration, supplier func() (T, error)) (result T, err error) {
	value, err := cache.Get(key)
	if err == nil {
		return serializeutil.Deserialize[T](value.([]byte))
	}
	if err != gcache.KeyNotFoundError {
		return result, err
	}
	result, err = supplier()
	if err != nil {
		return result, err
	}
	return result, cache.SetWithExpire(key, serializeutil.Serialize(result), expire)
}
