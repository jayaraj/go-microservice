package cache

import (
	"errors"
	"go-microservice/server"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	instance        *cacheService
	ErrCacheMiss    = errors.New("key not found")
	ErrNotStored    = errors.New("not stored")
	ErrInvalidValue = errors.New("invalid value")
)

const (
	DefaultExpiryTime  = time.Duration(0)
	ForEverNeverExpiry = time.Duration(-1)
)

type cacheService struct {
	cache Cache
}

func init() {
	instance = &cacheService{}
	server.RegisterService(instance, server.Low)
}

func (c *cacheService) Init() (err error) {
	defaultExpiration := time.Hour
	viper.SetDefault("cache", "inmemory")
	switch viper.Get("cache") {
	case "redis":
		{
			viper.SetDefault("redis", "")
			hosts := strings.Split(viper.Get("redis").(string), ",")
			if len(hosts) == 0 || len(hosts) > 1 {
				log.Error("None or more than one host configured")
				instance.cache = newInMemoryCache(defaultExpiration)
			} else {
				viper.SetDefault("redispassword", "")
				password := viper.Get("redispassword").(string)
				log.Infof("Initialised redis hosts : %s", hosts[0])
				instance.cache = newRedisCache(hosts[0], password, defaultExpiration)
			}
		}
	case "memcache":
		{
			viper.SetDefault("memcache", "")
			hosts := strings.Split(viper.Get("memcache").(string), ",")
			if len(hosts) == 0 {
				log.Error("No host configured")
				instance.cache = newInMemoryCache(defaultExpiration)
			} else {
				log.Infof("Initialised memcache hosts : %v", hosts)
				instance.cache = newMemcachedCache(hosts, defaultExpiration)
			}
		}
	default:
		{
			log.Infof("Initialised in-memory cache")
			instance.cache = newInMemoryCache(defaultExpiration)
		}
	}
	return nil
}

func (c *cacheService) OnConfig() {
	c.Init()
}

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
	return instance.cache.Get(key, ptrValue)
}

func GetMulti(keys ...string) (Getter, error) {
	return instance.cache.GetMulti(keys...)
}

func Delete(key string) error {
	return instance.cache.Delete(key)
}

func Increment(key string, n uint64) (newValue uint64, err error) {
	return instance.cache.Increment(key, n)
}

func Decrement(key string, n uint64) (newValue uint64, err error) {
	return instance.cache.Decrement(key, n)
}

func Flush() error {
	return instance.cache.Flush()
}

func Set(key string, value interface{}, expires time.Duration) error {
	return instance.cache.Set(key, value, expires)
}

func Add(key string, value interface{}, expires time.Duration) error {
	return instance.cache.Add(key, value, expires)
}

func Replace(key string, value interface{}, expires time.Duration) error {
	return instance.cache.Replace(key, value, expires)
}
