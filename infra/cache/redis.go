package cache

import (
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type redisCache struct {
	pool              *redis.Pool
	defaultExpiration time.Duration
}

type redisItemMapGetter map[string][]byte

func init() {
	// defaultExpiration := time.Hour
	// remoteCache = &redisCache{
	// 	defaultExpiration: defaultExpiration,
	// }
	// server.RegisterService(remoteCache.(*redisCache), server.Low)
}

func (c *redisCache) Init() (err error) {
	viper.SetDefault("redis", "")
	hosts := strings.Split(viper.Get("redis").(string), ",")
	if len(hosts) == 0 || len(hosts) > 1 {
		log.Fatalln("None or more than one host configured")
	}
	viper.SetDefault("redispassword", "")
	password := viper.Get("redispassword").(string)
	log.Infof("Initialised redis hosts : %s", hosts[0])
	c.pool = newRedisCache(hosts[0], password, c.defaultExpiration)
	return nil
}

func (c *redisCache) OnConfig() {
}

func newRedisCache(host string, password string, defaultExpiration time.Duration) *redis.Pool {

	return &redis.Pool{
		MaxIdle:     5,
		MaxActive:   0,
		IdleTimeout: time.Duration(240) * time.Second,
		Dial: func() (redis.Conn, error) {
			protocol := "tcp"
			toc := time.Millisecond * time.Duration(10000)
			tor := time.Millisecond * time.Duration(5000)
			tow := time.Millisecond * time.Duration(5000)
			c, err := redis.Dial(protocol, host,
				redis.DialConnectTimeout(toc),
				redis.DialReadTimeout(tor),
				redis.DialWriteTimeout(tow))
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err = c.Do("AUTH", password); err != nil {
					_ = c.Close()
					return nil, err
				}
			} else {
				// check with PING
				if _, err = c.Do("PING"); err != nil {
					_ = c.Close()
					return nil, err
				}
			}
			return c, err
		},
		// custom connection test method
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (c *redisCache) Set(key string, value interface{}, expires time.Duration) error {
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	return c.invoke(conn.Do, key, value, expires)
}

func (c *redisCache) Add(key string, value interface{}, expires time.Duration) error {
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	existed, err := exists(conn, key)
	if err != nil {
		return err
	} else if existed {
		return ErrNotStored
	}
	return c.invoke(conn.Do, key, value, expires)
}

func (c *redisCache) Replace(key string, value interface{}, expires time.Duration) error {
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	existed, err := exists(conn, key)
	if err != nil {
		return err
	} else if !existed {
		return ErrNotStored
	}

	err = c.invoke(conn.Do, key, value, expires)
	if value == nil {
		return ErrNotStored
	}
	return err
}

func (c *redisCache) Get(key string, ptrValue interface{}) error {
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	raw, err := conn.Do("GET", key)
	if err != nil {
		return err
	} else if raw == nil {
		return ErrCacheMiss
	}
	item, err := redis.Bytes(raw, err)
	if err != nil {
		return err
	}
	return deserialize(item, ptrValue)
}

func generalizeStringSlice(strs []string) []interface{} {
	ret := make([]interface{}, len(strs))
	for i, str := range strs {
		ret[i] = str
	}
	return ret
}

func (c *redisCache) GetMulti(keys ...string) (Getter, error) {
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	items, err := redis.Values(conn.Do("MGET", generalizeStringSlice(keys)...))
	if err != nil {
		return nil, err
	} else if items == nil {
		return nil, ErrCacheMiss
	}

	m := make(map[string][]byte)
	for i, key := range keys {
		m[key] = nil
		if i < len(items) && items[i] != nil {
			s, ok := items[i].([]byte)
			if ok {
				m[key] = s
			}
		}
	}
	return redisItemMapGetter(m), nil
}

func exists(conn redis.Conn, key string) (bool, error) {
	return redis.Bool(conn.Do("EXISTS", key))
}

func (c *redisCache) Delete(key string) error {
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	existed, err := redis.Bool(conn.Do("DEL", key))
	if err == nil && !existed {
		err = ErrCacheMiss
	}
	return err
}

func (c *redisCache) Increment(key string, delta uint64) (uint64, error) {
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	// Check for existence *before* increment as per the cache contract.
	// redis will auto create the key, and we don't want that. Since we need to do increment
	// ourselves instead of natively via INCRBY (redis doesn't support wrapping), we get the value
	// and do the exists check this way to minimize calls to Redis
	val, err := conn.Do("GET", key)
	if err != nil {
		return 0, err
	} else if val == nil {
		return 0, ErrCacheMiss
	}
	currentVal, err := redis.Int64(val, nil)
	if err != nil {
		return 0, err
	}
	sum := currentVal + int64(delta)
	_, err = conn.Do("SET", key, sum)
	if err != nil {
		return 0, err
	}
	return uint64(sum), nil
}

func (c *redisCache) Decrement(key string, delta uint64) (newValue uint64, err error) {
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	// Check for existence *before* increment as per the cache contract.
	// redis will auto create the key, and we don't want that, hence the exists call
	existed, err := exists(conn, key)
	if err != nil {
		return 0, err
	} else if !existed {
		return 0, ErrCacheMiss
	}
	// Decrement contract says you can only go to 0
	// so we go fetch the value and if the delta is greater than the amount,
	// 0 out the value
	currentVal, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		return 0, err
	}
	if delta > uint64(currentVal) {
		var tempint int64
		tempint, err = redis.Int64(conn.Do("DECRBY", key, currentVal))
		return uint64(tempint), err
	}
	tempint, err := redis.Int64(conn.Do("DECRBY", key, delta))
	return uint64(tempint), err
}

func (c *redisCache) Flush() error {
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	_, err := conn.Do("FLUSHALL")
	return err
}

func (c *redisCache) invoke(f func(string, ...interface{}) (interface{}, error),
	key string, value interface{}, expires time.Duration) error {

	switch expires {
	case DefaultExpiryTime:
		expires = c.defaultExpiration
	case ForEverNeverExpiry:
		expires = time.Duration(0)
	}

	b, err := serialize(value)
	if err != nil {
		return err
	}
	conn := c.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	if expires > 0 {
		_, err = f("SETEX", key, int32(expires/time.Second), b)
		return err
	}
	_, err = f("SET", key, b)
	return err
}

func (g redisItemMapGetter) Get(key string, ptrValue interface{}) error {
	item, ok := g[key]
	if !ok {
		return ErrCacheMiss
	}
	return deserialize(item, ptrValue)
}
