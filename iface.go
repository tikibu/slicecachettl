package slicecachettl

import "time"

type Timeable interface {
	GetTS() time.Time
}

type ExpirationHandler func(key interface{}, arr []Timeable)
type CacheTtl interface {
	Append(key interface{}, value Timeable) error
	Get(key interface{}) (ret []Timeable, ok bool)
}

type CacheTtlFactory interface {
	WithExpirationHandler(onExpiration ExpirationHandler, ttl time.Duration, maxresolution time.Duration) CacheTtl
	Simple(ttl time.Duration, maxresolution time.Duration) CacheTtl
	Custom(onExpiration ExpirationHandler,
		ttl time.Duration,
		expirationResolution chan Timeable,
		defaultSliceSize int,
		expirationChanSize int) CacheTtl
}
