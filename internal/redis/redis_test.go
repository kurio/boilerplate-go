package redis_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	_redis "github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	goboilerplate "github.com/kurio/boilerplate-go"
	"github.com/kurio/boilerplate-go/internal/redis"
)

type redisTestSuite struct {
	suite.Suite
	redisClient _redis.UniversalClient
}

func (s *redisTestSuite) SetupSuite() {
	addr := os.Getenv("REDIS_TEST")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	redisClusterStr := os.Getenv("REDIS_CLUSTER")
	if redisClusterStr == "" {
		logrus.Warnf("REDIS_CLUSTER is not set, assume to be false")
		redisClusterStr = "false"
	}

	redisCluster, err := strconv.ParseBool(redisClusterStr)
	if err != nil {
		logrus.Warnf("REDIS_CLUSTER is not a valid boolean, set to false")
		redisCluster = false
	}

	var redisClient _redis.UniversalClient
	if redisCluster {
		redisClient = _redis.NewClusterClient(&_redis.ClusterOptions{
			Addrs:        strings.Split(addr, ","),
			DialTimeout:  2 * time.Second,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     5,
			PoolTimeout:  30 * time.Second,
		})
	} else {
		redisClient = _redis.NewClient(&_redis.Options{
			Addr:         addr,
			DialTimeout:  2 * time.Second,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     5,
			PoolTimeout:  30 * time.Second,
		})
	}

	require.NoError(s.T(), tryPing(redisClient))

	s.redisClient = redisClient
}

func tryPing(db _redis.UniversalClient) (err error) {
	const maxRetry = 30
	const interval = 1 * time.Second
	maxAttempts := maxRetry + 1
	for i := 0; i < maxAttempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = db.Ping(ctx).Err()
		if err == nil {
			return
		}

		time.Sleep(interval)
	}
	return
}

func (s *redisTestSuite) TearDownSuite() {
	s.redisClient.FlushDB(context.Background())

	require.NoError(s.T(), s.redisClient.Close())
}

func TestRedisCacher(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped for short testing")
	}

	suite.Run(t, new(redisTestSuite))
}

func (s *redisTestSuite) TestGet() {
	t := s.T()
	ctx := context.Background()

	key := "some-key"
	value := "some-value"

	cacher := redis.NewRedisCacher(s.redisClient, goboilerplate.ExpiryConf{}, "test")

	t.Run("not found", func(t *testing.T) {
		_, err := cacher.Get(ctx, key)
		require.Error(t, err)
		require.EqualError(t, errors.Cause(err), goboilerplate.ErrNotFound.Error())
	})

	t.Run("found", func(t *testing.T) {
		err := cacher.Set(
			ctx,
			key,
			value,
			goboilerplate.DurationShort,
		)
		require.NoError(t, err)

		res, err := cacher.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, value, res)
	})
}

func (s *redisTestSuite) TestSet() {
	t := s.T()
	ctx := context.Background()

	key := "some-key"
	value := "some-value"

	t.Run("without key prefix", func(t *testing.T) {
		cacher := redis.NewRedisCacher(s.redisClient, goboilerplate.ExpiryConf{}, "")

		err := cacher.Set(ctx, key, value, goboilerplate.DurationShort)
		require.NoError(t, err)

		res, err := s.redisClient.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, value, res)
	})

	t.Run("with key prefix", func(t *testing.T) {
		cacher := redis.NewRedisCacher(s.redisClient, goboilerplate.ExpiryConf{}, "test")

		err := cacher.Set(ctx, key, value, goboilerplate.DurationShort)
		require.NoError(t, err)

		res, err := s.redisClient.Get(ctx, fmt.Sprintf("test###%s", key)).Result()
		require.NoError(t, err)
		require.Equal(t, value, res)
	})
}

func (s *redisTestSuite) TestDel() {
	t := s.T()
	ctx := context.Background()

	key := "some-key"
	value := "some-value"

	cacher := redis.NewRedisCacher(s.redisClient, goboilerplate.ExpiryConf{}, "test")

	t.Run("success", func(t *testing.T) {
		err := cacher.Set(
			ctx,
			key,
			value,
			goboilerplate.DurationShort,
		)
		require.NoError(t, err)

		err = cacher.Del(ctx, key)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		err := cacher.Del(ctx, key)
		require.Error(t, err)
		require.EqualError(t, errors.Cause(err), goboilerplate.ErrNotFound.Error())
	})
}

func (s *redisTestSuite) TestFlush() {
	t := s.T()
	ctx := context.Background()

	cacher := redis.NewRedisCacher(s.redisClient, goboilerplate.ExpiryConf{}, "test")

	require.NoError(t, cacher.Set(ctx, "k1", "v1", goboilerplate.DurationShort))
	require.NoError(t, cacher.Set(ctx, "k2", "v2", goboilerplate.DurationShort))
	require.NoError(t, cacher.Set(ctx, "k3", "v3", goboilerplate.DurationShort))

	var err error
	_, err = cacher.Get(ctx, "k1")
	require.NoError(t, err)
	_, err = cacher.Get(ctx, "k2")
	require.NoError(t, err)
	_, err = cacher.Get(ctx, "k3")
	require.NoError(t, err)

	require.NoError(t, cacher.Flush(ctx))

	_, err = cacher.Get(ctx, "k1")
	require.Error(t, err)
	require.EqualError(t, errors.Cause(err), goboilerplate.ErrNotFound.Error())

	_, err = cacher.Get(ctx, "k2")
	require.Error(t, err)
	require.EqualError(t, errors.Cause(err), goboilerplate.ErrNotFound.Error())

	_, err = cacher.Get(ctx, "k3")
	require.Error(t, err)
	require.EqualError(t, errors.Cause(err), goboilerplate.ErrNotFound.Error())
}
