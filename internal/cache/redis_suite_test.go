package cache_test

import (
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	dBPingMaxRetry      = 30
	dBPingRetryInterval = 1 * time.Second
)

type redisSuite struct {
	suite.Suite
	DB *redis.Client
}

func (s *redisSuite) SetupSuite() {
	addr := os.Getenv("REDIS_TEST")
	if addr == "" {
		addr = "localhost:6379"
	}

	db := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	if err := tryPing(db, dBPingMaxRetry, dBPingRetryInterval); err != nil {
		require.NoError(s.T(), err)
	}

	s.DB = db
}

func tryPing(db *redis.Client, maxRetry int, interval time.Duration) (err error) {
	maxAttempts := maxRetry + 1
	for i := 0; i < maxAttempts; i++ {
		if _, err = db.Ping().Result(); err != nil {
			time.Sleep(interval)
			continue
		}
		return nil
	}
	return err
}

func (s *redisSuite) TearDownSuite() {
	s.DB.FlushDB()

	if err := s.DB.Close(); err != nil {
		require.NoError(s.T(), err)
	}
}
