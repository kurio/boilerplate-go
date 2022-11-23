package goboilerplate

import (
	"context"
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

// Default duration for cache expiry
const (
	defaultDurationShort = 1 * time.Minute
	defaultDurationLong  = 1 * time.Hour
)

// ExpiryConf defines the configuration for each expiry duration.
type ExpiryConf map[ExpiryDuration]time.Duration

// Set the default duration if it's not already set.
func (c ExpiryConf) Set() {
	if _, ok := c[DurationShort]; !ok {
		c[DurationShort] = defaultDurationShort
	}
	if _, ok := c[DurationLong]; !ok {
		c[DurationLong] = defaultDurationLong
	}
}

// Cacher is the interface of a data cacher.
type Cacher interface {
	Get(ctx context.Context, key string) (value string, err error)
	Set(ctx context.Context, key string, value string, expiration ExpiryDuration) error
	Del(ctx context.Context, key string) error
	Flush(ctx context.Context) error
}
