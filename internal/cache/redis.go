package cache

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-redis/redis"
)

type redisClient struct {
	conn       *redis.Client
	expiryConf ExpiryConf
}

// NewRedisClient is a constructor for caching using redis
func NewRedisClient(redisCon *redis.Client, expiryConf ExpiryConf) DataCacher {
	return redisClient{
		conn:       redisCon,
		expiryConf: expiryConf,
	}
}

func (s redisClient) Get(key string, data interface{}) (err error) {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("%w", ErrConstraint)
	}

	b, err := s.conn.Get(key).Bytes()
	if err != nil {
		if len(b) == 0 {
			err = fmt.Errorf("%w", ErrNotFound)
			return
		}

		return
	}

	err = json.Unmarshal(b, &data)
	return
}

func (s redisClient) Set(key string, data interface{}, expiration ExpiryDuration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := s.conn.Set(key, b, s.expiryConf[expiration]).Err(); err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}
