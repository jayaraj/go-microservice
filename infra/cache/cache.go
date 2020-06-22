package cache

import (
	"errors"
	"time"
)

var (
	internalCache   Cache
	remoteCache     Cache
	ErrCacheMiss    = errors.New("key not found")
	ErrNotStored    = errors.New("not stored")
	ErrInvalidValue = errors.New("invalid value")
)

const (
	DefaultExpiryTime  = time.Duration(0)
	ForEverNeverExpiry = time.Duration(-1)
)

type Getter interface {
	Get(key string, ptrValue interface{}) error
}

type Cache interface {
	Getter
	Set(key string, value interface{}, expires time.Duration) error
	GetMulti(keys ...string) (Getter, error)
	Delete(key string) error
	Add(key string, value interface{}, expires time.Duration) error
	Replace(key string, value interface{}, expires time.Duration) error
	Increment(key string, n uint64) (newValue uint64, err error)
	Decrement(key string, n uint64) (newValue uint64, err error)
	Flush() error
}

// Get the Content associated with key from the cache.
// Returns an  error if not found
func Get(remote bool, key string, ptrValue interface{}) error {
	if remote {
		return remoteCache.Get(key, ptrValue)
	}
	return internalCache.Get(key, ptrValue)
}

// Get the content associated multiple keys at once.  On success, the caller
// may decode the values one at a time from the returned Getter.
func GetMulti(remote bool, keys ...string) (Getter, error) {
	if remote {
		return remoteCache.GetMulti(keys...)
	}
	return internalCache.GetMulti(keys...)
}

// Delete the given key from the cache.
func Delete(remote bool, key string) error {
	if remote {
		return remoteCache.Delete(key)
	}
	return internalCache.Delete(key)
}

// Increment the value stored at the given key by the given amount.
// The value silently wraps around upon exceeding the uint64 range.
func Increment(remote bool, key string, n uint64) (newValue uint64, err error) {
	if remote {
		return remoteCache.Increment(key, n)
	}
	return internalCache.Increment(key, n)
}

// Decrement the value stored at the given key by the given amount.
// The value is capped at 0 on underflow, with no error returned.
func Decrement(remote bool, key string, n uint64) (newValue uint64, err error) {
	if remote {
		return remoteCache.Decrement(key, n)
	}
	return internalCache.Decrement(key, n)
}

// Expire all cache entries immediately.
// This is not implemented for the memcached cache (intentionally).
// Returns an implementation specific error if the operation failed.
func Flush(remote bool) error {
	if remote {
		return remoteCache.Flush()
	}
	return internalCache.Flush()
}

// Set the given key/value in the cache, overwriting any existing value
// associated with that key.  Keys may be at most 250 bytes in length.
func Set(remote bool, key string, value interface{}, expires time.Duration) error {
	if remote {
		return remoteCache.Set(key, value, expires)
	}
	return internalCache.Set(key, value, expires)
}

// Add the given key/value to the cache ONLY IF the key does not already exist.
func Add(remote bool, key string, value interface{}, expires time.Duration) error {
	if remote {
		return remoteCache.Add(key, value, expires)
	}
	return internalCache.Add(key, value, expires)
}

// Set the given key/value in the cache ONLY IF the key already exists.
func Replace(remote bool, key string, value interface{}, expires time.Duration) error {
	if remote {
		return remoteCache.Replace(key, value, expires)
	}
	return internalCache.Replace(key, value, expires)
}
