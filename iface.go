package slicecachettl

import "time"

type Timeable interface {
	GetTS() time.Time
}

type ExpirationHandler func(key interface{}, arr []Timeable)
type InstrumentationHandler func(mapLen, listLen, chanLen int)

type CacheTtl interface {
	Append(key interface{}, value Timeable) error
	CheckAndLock(key interface{}, value Timeable) bool
	Get(key interface{}) (ret []Timeable, ok bool)
	ExpireAll() int // Returns number of expired items
	ExpireCustom(d time.Duration) int
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
