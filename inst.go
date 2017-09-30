package slicecachettl

import "time"

var Factory = &Cache{}

func WithExpirationHandler(onExpiration ExpirationHandler, ttl time.Duration, maxresolution time.Duration) CacheTtl {
	return Factory.WithExpirationHandler(onExpiration, ttl, maxresolution)
}

func Simple(ttl time.Duration, maxresolution time.Duration) CacheTtl {
	return Factory.Simple(ttl, maxresolution)
}

func Custom(onExpiration ExpirationHandler,
	ttl time.Duration,
	expirationResolution chan Timeable,
	defaultSliceSize int,
	expirationChanSize int) CacheTtl {
	return Factory.Custom(onExpiration, ttl, expirationResolution, defaultSliceSize, expirationChanSize)
}
