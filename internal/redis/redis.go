package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"

	goboilerplate "github.com/kurio/boilerplate-go"
)

type redisCacher struct {
	redisClient redis.UniversalClient
	expiryConf  goboilerplate.ExpiryConf
	keyPrefix   string
}

// NewRedisCacher is a constructor for caching using redis
func NewRedisCacher(redisClient redis.UniversalClient, expiryConf goboilerplate.ExpiryConf, keyPrefix string) goboilerplate.Cacher {
	expiryConf.Set()

	return redisCacher{
		redisClient: redisClient,
		expiryConf:  expiryConf,
		keyPrefix:   keyPrefix,
	}
}

func (c redisCacher) getKey(key string) string {
	if c.keyPrefix == "" {
		return key
	}
	return fmt.Sprintf("%s###%s", c.keyPrefix, key)
}

func (c redisCacher) Get(ctx context.Context, key string) (value string, err error) {
	value, err = c.redisClient.Get(ctx, c.getKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			err = goboilerplate.ErrNotFound
			return
		}
		err = errors.Wrap(err, "error getting data from redis")
		return
	}

	return
}

func (c redisCacher) Set(ctx context.Context, key string, value string, expiration goboilerplate.ExpiryDuration) (err error) {
	err = c.redisClient.Set(
		ctx,
		c.getKey(key),
		value,
		c.expiryConf[expiration],
	).Err()
	if err != nil {
		err = errors.Wrap(err, "error setting data to redis")
		return
	}

	return
}

func (c redisCacher) Del(ctx context.Context, key string) (err error) {
	var count int64
	count, err = c.redisClient.Del(ctx, c.getKey(key)).Result()
	if err != nil {
		err = errors.Wrap(err, "error deleting data from redis")
		return
	}
	if count == 0 {
		err = goboilerplate.ErrNotFound
		return
	}
	return
}

func (s redisCacher) Flush(ctx context.Context) (err error) {
	err = s.redisClient.FlushDB(ctx).Err()
	if err != nil {
		err = errors.Wrap(err, "error flushing redis cache")
		return
	}

	return
}
