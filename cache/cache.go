package cache

import (
	"errors"
	"time"
)

var (
	instance        Cache
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

	// Set the given key/value in the cache, overwriting any existing value
	// associated with that key.  Keys may be at most 250 bytes in length.
	//
	// Returns:
	//   - nil on success
	//   - an implementation specific error otherwise
	Set(key string, value interface{}, expires time.Duration) error

	// Get the content associated multiple keys at once.  On success, the caller
	// may decode the values one at a time from the returned Getter.
	//
	// Returns:
	//   - the value getter, and a nil error if the operation completed.
	//   - an implementation specific error otherwise
	GetMulti(keys ...string) (Getter, error)

	// Delete the given key from the cache.
	//
	// Returns:
	//   - nil on a successful delete
	//   - ErrCacheMiss if the value was not in the cache
	//   - an implementation specific error otherwise
	Delete(key string) error

	// Add the given key/value to the cache ONLY IF the key does not already exist.
	//
	// Returns:
	//   - nil if the value was added to the cache
	//   - ErrNotStored if the key was already present in the cache
	//   - an implementation-specific error otherwise
	Add(key string, value interface{}, expires time.Duration) error

	// Set the given key/value in the cache ONLY IF the key already exists.
	//
	// Returns:
	//   - nil if the value was replaced
	//   - ErrNotStored if the key does not exist in the cache
	//   - an implementation specific error otherwise
	Replace(key string, value interface{}, expires time.Duration) error

	// Increment the value stored at the given key by the given amount.
	// The value silently wraps around upon exceeding the uint64 range.
	//
	// Returns the new counter value if the operation was successful, or:
	//   - ErrCacheMiss if the key was not found in the cache
	//   - an implementation specific error otherwise
	Increment(key string, n uint64) (newValue uint64, err error)

	// Decrement the value stored at the given key by the given amount.
	// The value is capped at 0 on underflow, with no error returned.
	//
	// Returns the new counter value if the operation was successful, or:
	//   - ErrCacheMiss if the key was not found in the cache
	//   - an implementation specific error otherwise
	Decrement(key string, n uint64) (newValue uint64, err error)

	// Expire all cache entries immediately.
	// This is not implemented for the memcached cache (intentionally).
	// Returns an implementation specific error if the operation failed.
	Flush() error
}

func Get(key string, ptrValue interface{}) error {
	return instance.Get(key, ptrValue)
}

func GetMulti(keys ...string) (Getter, error) {
	return instance.GetMulti(keys...)
}

func Delete(key string) error {
	return instance.Delete(key)
}

func Increment(key string, n uint64) (newValue uint64, err error) {
	return instance.Increment(key, n)
}

func Decrement(key string, n uint64) (newValue uint64, err error) {
	return instance.Decrement(key, n)
}

func Flush() error {
	return instance.Flush()
}

func Set(key string, value interface{}, expires time.Duration) error {
	return instance.Set(key, value, expires)
}

func Add(key string, value interface{}, expires time.Duration) error {
	return instance.Add(key, value, expires)
}

func Replace(key string, value interface{}, expires time.Duration) error {
	return instance.Replace(key, value, expires)
}
