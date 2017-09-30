# slicecachettl
Simple performant ttl cache, w/ constant-time insertions and expirations.

It's for a vary narrow use case to be a part of an online data joiner. It can be used for online joining labels (click/no click) to training inputs.

For context, you can read about joiner in e.g.
> X. He, J. Pan, O. Jin, T. Xu, B. Liu, T. Xu, Y. Shi, A. Atallah, R. Herbrich, S. Bowers, and J. Q. n. Candela.
> Practical lessons from predicting clicks on ads at facebook.
> In Proceedings of the Eighth International Workshop on
> Data Mining for Online Advertising, ADKDD?14, 2014

How to use
=================

The keys are of type `interface{}` and need to be compliant with requirements for `map` type.

The values need to satisfy interface Timeable:

```
type Timeable interface {
	GetTS() time.Time
}
```

Example (Simple Append/Get):
--------

```

	cache := Simple(time.Millisecond*100, time.Millisecond*10)
	start := time.Now()
	const KEYS = 10
	for key := 0; key <= KEYS; key++ {
		cache.Append(key, &TimeableTest{K: int(key), V: key, Ts: start})
	}

	v, ok := cache.Get(1)

```


Example (with expiration handler)
-----
```
	cache := WithExpirationHandler(func(key interface{}, arr []Timeable) {
	    fmt.Println("hello")
	}, time.Millisecond*100, time.Millisecond*10)

	cache.Append(key, &TimeableTest{K: int(key), V: key, Ts: start})

```
