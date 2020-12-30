package cache

import (
	"errors"
	"time"
)

// ExpiryDuration is used to define the expiry duration when caching some data.
type ExpiryDuration uint8

const (
	// DurationShort is used for caching some data that might change in
	// a short time.
	// It could also be used when we're not able to invalidate the cache.
	DurationShort ExpiryDuration = iota

	// DurationLong is typically used for caching some data that we own.
	// This way, we could invalidate the cache any time the data changes.
	DurationLong
)

// ExpiryConf defines the configuration for each expiry duration.
type ExpiryConf map[ExpiryDuration]time.Duration

var (
	// ErrConstraint represents a custom error for a contstraint things.
	ErrConstraint = errors.New("invalid passed argument")

	// ErrNotFound represents a custom error for a not exists key
	ErrNotFound = errors.New("key is not found")
)

// DataCacher is the interface of a data cacher.
type DataCacher interface {
	Get(key string, data interface{}) error
	Set(key string, data interface{}, expiration ExpiryDuration) error
}
