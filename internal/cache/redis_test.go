package cache_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/kurio/boilerplate-go/internal/cache"
)

type object1 struct {
	Str string
	Num int
}

type object2 struct {
	Title  string
	Images []string
}

var expirationTime = map[cache.ExpiryDuration]time.Duration{
	cache.DurationShort: 5 * time.Minute,
	cache.DurationLong:  10 * time.Minute,
}

type redisTestSuite struct {
	redisSuite
}

func TestRedisTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped for short testing")
	}

	suite.Run(t, new(redisTestSuite))
}

func (s *redisTestSuite) TestRedisCacheGet() {
	var target1 object1
	var target2 object2

	tests := []struct {
		testName       string
		key            string
		target         interface{}
		shouldCache    bool
		expectedResult interface{}
		expectedError  error
	}{
		{
			testName:       "success 1",
			key:            "myKey1",
			target:         &target1,
			shouldCache:    true,
			expectedResult: object1{Str: "mystring", Num: 42},
			expectedError:  nil,
		},
		{
			testName:    "success 2",
			key:         "myKey2",
			target:      &target2,
			shouldCache: true,
			expectedResult: object2{
				Title: "Gempa Kroasia : 7 Tewas dan Puluhan Korban Lain tertimpa Reruntuhan",
				Images: []string{
					"https://kurio-img.kurioapps.com/20/12/30/6afa2c26-1d26-4165-ba92-9aa50d31b698.jpg",
				},
			},
			expectedError: nil,
		},
		{
			testName:       "data not found",
			key:            "404",
			target:         &target2,
			shouldCache:    false,
			expectedResult: object2{},
			expectedError:  cache.ErrNotFound,
		},
	}

	for _, test := range tests {
		s.T().Run(test.testName, func(t *testing.T) {
			if test.shouldCache {
				byteObj, err := json.Marshal(test.expectedResult)
				require.NoError(s.T(), err)

				err = s.DB.Set(test.key, byteObj, expirationTime[cache.DurationLong]).Err()
				require.NoError(t, err)
			}

			cacher := cache.NewRedisClient(s.DB, cache.ExpiryConf{})

			rtarget := reflect.ValueOf(test.target)
			rtarget.Elem().Set(reflect.Zero(reflect.TypeOf(test.target).Elem()))

			err := cacher.Get(test.key, test.target)

			if test.expectedError != nil {
				require.EqualError(s.T(), err, test.expectedError.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, rtarget.Elem().Interface(), test.expectedResult)
		})
	}
}

func (s *redisTestSuite) TestRedisCacheSet() {
	key := "myKey"
	obj := object1{Str: "mystring", Num: 42}

	s.T().Run("success", func(t *testing.T) {
		repo := cache.NewRedisClient(s.DB, expirationTime)

		err := repo.Set(key, obj, cache.DurationShort)
		require.NoError(t, err)

		b, err := s.DB.Get(key).Bytes()
		require.NoError(t, err)

		var data object1
		err = json.Unmarshal(b, &data)

		require.NoError(t, err)
		require.Equal(t, data, obj)
	})
}
