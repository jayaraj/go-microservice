package cache

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

type inMemoryCache struct {
	cache cache.Cache
	mu    sync.RWMutex
}

func init() {
	//TODO
	// defaultExpiration := time.Hour
	// instance = &inMemoryCache{
	// 	cache: *cache.New(defaultExpiration, time.Minute),
	// 	mu:    sync.RWMutex{},
	// }
	// server.RegisterService(instance.(*inMemoryCache), server.Low)
}

func (c *inMemoryCache) Init() (err error) {
	//Do nothing
	return nil
}

func (c *inMemoryCache) OnConfig() {
	//Do nothing
}

func newInMemoryCache(defaultExpiration time.Duration) *inMemoryCache {
	return &inMemoryCache{
		cache: *cache.New(defaultExpiration, time.Minute),
		mu:    sync.RWMutex{},
	}
}

func (c *inMemoryCache) Get(key string, ptrValue interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, found := c.cache.Get(key)
	if !found {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(ptrValue)
	if v.Type().Kind() == reflect.Ptr && v.Elem().CanSet() {
		v.Elem().Set(reflect.ValueOf(value))
		return nil
	}

	err := fmt.Errorf("attempt to get %s, but can not set value %v", key, v)
	log.Error(err.Error())
	return err
}

func (c *inMemoryCache) GetMulti(keys ...string) (Getter, error) {
	return c, nil
}

func (c *inMemoryCache) Set(key string, value interface{}, expires time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// NOTE: go-cache understands the values of DefaultExpiryTime and ForEverNeverExpiry
	c.cache.Set(key, value, expires)
	return nil
}

func (c *inMemoryCache) Add(key string, value interface{}, expires time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.cache.Add(key, value, expires)
	if err != nil {
		return ErrNotStored
	}
	return err
}

func (c *inMemoryCache) Replace(key string, value interface{}, expires time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.cache.Replace(key, value, expires); err != nil {
		return ErrNotStored
	}
	return nil
}

func (c *inMemoryCache) Delete(key string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, found := c.cache.Get(key); !found {
		return ErrCacheMiss
	}
	c.cache.Delete(key)
	return nil
}

func (c *inMemoryCache) Increment(key string, n uint64) (newValue uint64, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, found := c.cache.Get(key); !found {
		return 0, ErrCacheMiss
	}
	if err = c.cache.Increment(key, int64(n)); err != nil {
		return
	}

	return c.convertTypeToUint64(key)
}

func (c *inMemoryCache) Decrement(key string, n uint64) (newValue uint64, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if nv, err := c.convertTypeToUint64(key); err != nil {
		return 0, err
	} else {
		// Stop from going below zero
		if n > nv {
			n = nv
		}
	}
	if err = c.cache.Decrement(key, int64(n)); err != nil {
		return
	}

	return c.convertTypeToUint64(key)
}

func (c *inMemoryCache) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Flush()
	return nil
}

func (c *inMemoryCache) convertTypeToUint64(key string) (newValue uint64, err error) {
	v, found := c.cache.Get(key)
	if !found {
		return newValue, ErrCacheMiss
	}

	switch v.(type) {
	case int:
		newValue = uint64(v.(int))
	case int8:
		newValue = uint64(v.(int8))
	case int16:
		newValue = uint64(v.(int16))
	case int32:
		newValue = uint64(v.(int32))
	case int64:
		newValue = uint64(v.(int64))
	case uint:
		newValue = uint64(v.(uint))
	case uintptr:
		newValue = uint64(v.(uintptr))
	case uint8:
		newValue = uint64(v.(uint8))
	case uint16:
		newValue = uint64(v.(uint16))
	case uint32:
		newValue = uint64(v.(uint32))
	case uint64:
		newValue = uint64(v.(uint64))
	case float32:
		newValue = uint64(v.(float32))
	case float64:
		newValue = uint64(v.(float64))
	default:
		err = ErrInvalidValue
	}
	return
}
