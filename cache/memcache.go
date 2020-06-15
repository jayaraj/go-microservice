package cache

import (
	"errors"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	log "github.com/sirupsen/logrus"
)

type memcachedCache struct {
	*memcache.Client
	defaultExpiration time.Duration
}

type itemMapGetter map[string]*memcache.Item

func newMemcachedCache(hostList []string, defaultExpiration time.Duration) *memcachedCache {
	return &memcachedCache{
		memcache.New(hostList...),
		defaultExpiration,
	}
}

func (c *memcachedCache) Set(key string, value interface{}, expires time.Duration) error {
	return c.invoke((*memcache.Client).Set, key, value, expires)
}

func (c *memcachedCache) Add(key string, value interface{}, expires time.Duration) error {
	return c.invoke((*memcache.Client).Add, key, value, expires)
}

func (c *memcachedCache) Replace(key string, value interface{}, expires time.Duration) error {
	return c.invoke((*memcache.Client).Replace, key, value, expires)
}

func (c *memcachedCache) Get(key string, ptrValue interface{}) error {
	item, err := c.Client.Get(key)
	if err != nil {
		return convertMemcacheError(err)
	}
	return Deserialize(item.Value, ptrValue)
}

func (c *memcachedCache) GetMulti(keys ...string) (Getter, error) {
	items, err := c.Client.GetMulti(keys)
	if err != nil {
		return nil, convertMemcacheError(err)
	}
	return itemMapGetter(items), nil
}

func (c *memcachedCache) Delete(key string) error {
	return convertMemcacheError(c.Client.Delete(key))
}

func (c *memcachedCache) Increment(key string, delta uint64) (newValue uint64, err error) {
	newValue, err = c.Client.Increment(key, delta)
	return newValue, convertMemcacheError(err)
}

func (c *memcachedCache) Decrement(key string, delta uint64) (newValue uint64, err error) {
	newValue, err = c.Client.Decrement(key, delta)
	return newValue, convertMemcacheError(err)
}

func (c *memcachedCache) Flush() error {
	err := errors.New("Flush: can not flush memcached")
	log.Error(err.Error())
	return err
}

func (c *memcachedCache) invoke(f func(*memcache.Client, *memcache.Item) error,
	key string, value interface{}, expires time.Duration) error {

	switch expires {
	case DefaultExpiryTime:
		expires = c.defaultExpiration
	case ForEverNeverExpiry:
		expires = time.Duration(0)
	}

	b, err := Serialize(value)
	if err != nil {
		return err
	}
	return convertMemcacheError(f(c.Client, &memcache.Item{
		Key:        key,
		Value:      b,
		Expiration: int32(expires / time.Second),
	}))
}

func (g itemMapGetter) Get(key string, ptrValue interface{}) error {
	item, ok := g[key]
	if !ok {
		return ErrCacheMiss
	}

	return Deserialize(item.Value, ptrValue)
}

func convertMemcacheError(err error) error {
	switch err {
	case nil:
		return nil
	case memcache.ErrCacheMiss:
		return ErrCacheMiss
	case memcache.ErrNotStored:
		return ErrNotStored
	}

	log.WithFields(log.Fields{
		"error": err,
	}).Error("convertMemcacheError")
	return err
}
