package slicecachettl

import (
	"container/list"
	"sync"
	"time"
)

type MP map[interface{}][]Timeable
type SliceCacheTTL struct {
	expirations      list.List
	mp               MP
	defaultSliceSize int
	mu               sync.Mutex //read op is rare
	ttl              time.Duration
	onExpiration     ExpirationHandler
	expPool          chan time.Time
}

func (self *SliceCacheTTL) expire(now time.Time) (expired int) {
	for {
		self.mu.Lock()
		e := self.expirations.Back()
		if e == nil { //an empty expiration queue
			self.mu.Unlock()
			return expired
		}
		item := e.Value.(*Item)
		if item.Value.GetTS().Add(self.ttl).After(now) { // we deleted all till this moment
			self.mu.Unlock()
			return expired
		}

		self.expirations.Remove(e)
		expired += 1
		v, ok := self.mp[item.Key]
		if ok {
			delete(self.mp, item.Key)
			if self.onExpiration != nil {
				go self.onExpiration(item.Key, v)
			}
		} // if not ok, it have been deleted
		self.mu.Unlock()
	}
}

type Item struct {
	Key   interface{}
	Value Timeable
}

// This is a function that "Locks" a certain key
// for a number of seconds in expiration
func (self *SliceCacheTTL) CheckAndLock(key interface{}, value Timeable) bool {
	if value == nil {
		return false
	}
	self.expPool <- value.GetTS()
	if key == nil {
		return false
	}
	self.mu.Lock()
	_, ok := self.mp[key]
	ret := true
	if !ok{
		tm := make([]Timeable, 1, 1)
		tm[0] = value
		self.mp[key] = tm
		self.expirations.PushFront(&Item{Key: key, Value: value})
		ret = false
	}
	self.mu.Unlock()
	return ret
}

func (self *SliceCacheTTL) Append(key interface{}, value Timeable) error {
	if value == nil {
		return nil
	}
	self.expPool <- value.GetTS()
	if key == nil {
		return nil
	}
	self.mu.Lock()
	arr, ok := self.mp[key]
	if ok {
		self.mp[key] = append(arr, value)
	} else {
		sliceCapacity := self.defaultSliceSize
		if sliceCapacity < 1 {
			sliceCapacity = 1
		}
		tm := make([]Timeable, 1, sliceCapacity)
		tm[0] = value
		self.mp[key] = tm
		self.expirations.PushFront(&Item{Key: key, Value: value})
	}
	self.mu.Unlock()
	return nil
}

func (self *SliceCacheTTL) Get(key interface{}) (ret []Timeable, ok bool) {
	self.mu.Lock()
	ret, ok = self.mp[key]
	self.mu.Unlock()
	return
}

func (self *SliceCacheTTL) ExpireAll() int {
	return self.expire(time.Now().Add(self.ttl * 2))
}

func (self *SliceCacheTTL) ExpireCustom(d time.Duration) int {
	return self.expire(time.Now().Add(d))
}

type SimpleTime struct {
	Ts time.Time
}

func (self *SimpleTime) GetTS() time.Time {
	return self.Ts
}

func MakeMaxResolutionChan(maxresolution time.Duration) chan Timeable {
	tmchan := make(chan Timeable)
	go func() {
		tm := time.Tick(maxresolution)
		for {
			st := &SimpleTime{Ts: <-tm}
			tmchan <- st
		}
	}()
	return tmchan
}

type Cache struct{}

const DEFAULT_SLICE_SIZE = 20
const DEFAULT_EXPIRATION_CHAN_SIZE = 10000

func (self *Cache) WithExpirationHandler(onExpiration ExpirationHandler,
	ttl time.Duration,
	maxresolution time.Duration) CacheTtl {
	return self.Custom(onExpiration, nil, ttl, MakeMaxResolutionChan(maxresolution), DEFAULT_SLICE_SIZE, DEFAULT_EXPIRATION_CHAN_SIZE)
}

func (self *Cache) Simple(ttl time.Duration,
	maxresolution time.Duration) CacheTtl {
	return self.Custom(func(key interface{}, arr []Timeable) {}, nil, ttl,
		MakeMaxResolutionChan(maxresolution), DEFAULT_SLICE_SIZE, DEFAULT_EXPIRATION_CHAN_SIZE)
}

func (self *Cache) Custom(onExpiration ExpirationHandler, insHandler InstrumentationHandler, ttl time.Duration,
	expirationResolution chan Timeable, defaultSliceSize int, expirationChanSize int) CacheTtl {
	cache := &SliceCacheTTL{
		onExpiration:     onExpiration,
		ttl:              ttl,
		defaultSliceSize: defaultSliceSize,
		mp:               MP{},
		expPool:          make(chan time.Time, expirationChanSize),
	}
	go func() {
		for {
			timeable := <-expirationResolution
			cache.Append(nil, timeable)
		}
	}()
	go func() {
		for {
			ts := <-cache.expPool
			cache.expire(ts)
		}
	}()
	if insHandler != nil {
		go func() {
			for {
				time.Sleep(time.Second)
				insHandler(len(cache.mp), cache.expirations.Len(), len(cache.expPool))
			}
		}()
	}
	return cache
}
