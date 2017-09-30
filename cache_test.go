package slicecachettl

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type TimeableTest struct {
	K  int
	V  int
	Ts time.Time
}

func (self *TimeableTest) GetTS() time.Time {
	return self.Ts
}

func shuffle(src []*TimeableTest) []*TimeableTest {
	dest := make([]*TimeableTest, len(src))
	perm := rand.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}
	return dest
}

func testExpirations(t *testing.T, factory CacheTtlFactory) {
	start := time.Now().UTC()
	const KEYS int32 = 200
	var keys [KEYS]int
	var MS = 100
	var VALS_PER_MS = 100
	var n_expired int32
	var total_n_expired int32

	arr := make([]*TimeableTest, 0, VALS_PER_MS*MS)
	for i := 0; i < MS; i++ { //"milliseconds"
		for k := 0; k < VALS_PER_MS; k++ {
			key := rand.Int31n(KEYS)
			val := keys[key]
			keys[key] += 1
			arr = append(arr, &TimeableTest{K: int(key), V: val, Ts: start.Add(time.Millisecond * time.Duration(i))})
		}

	}

	z := MP{}
	cache := factory.Custom(func(key interface{}, arr []Timeable) {
		atomic.AddInt32(&n_expired, 1)
		for _, _ = range arr {
			atomic.AddInt32(&total_n_expired, 1)
			assert.Equal(t, arr, z[key])
		}
	},
		time.Millisecond*200,
		make(chan Timeable),
		20,
		1)
	//fmt.Println(len(arr))
	for _, a := range arr {
		cache.Append(a.K, a)
		z[a.K] = append(z[a.K], a)
	}

	assert.Equal(t, n_expired, int32(0), "Not expired, not deleted")

	cache.Append(nil, &TimeableTest{Ts: start.Add(time.Millisecond * 250)})
	time.Sleep(time.Millisecond * 300)

	assert.Equal(t, int32(KEYS), n_expired, "Not expired, not deleted")
	assert.Equal(t, int32(len(arr)), total_n_expired, "Not expired, not deleted")
}

func TestExpirations(t *testing.T) {
	testExpirations(t, &Cache{})
}

func TestInstantiation(t *testing.T) {
	simple := (&Cache{}).Simple(time.Second, time.Second)
	assert.NotNil(t, simple)
}

func TestWithExpirationHandler(t *testing.T) {
	var n_expired int32
	cache := WithExpirationHandler(func(key interface{}, arr []Timeable) {
		atomic.AddInt32(&n_expired, 1)
	},
		time.Millisecond*100, time.Millisecond*10)

	start := time.Now()
	const KEYS = 10
	for key := 0; key <= KEYS; key++ {
		cache.Append(key, &TimeableTest{K: int(key), V: key, Ts: start})
	}

	for key := 0; key <= KEYS; key++ {
		v, ok := cache.Get(key)
		assert.True(t, ok)
		assert.NotNil(t, v)
	}

	time.Sleep(time.Millisecond * 200)
	for key := 0; key <= KEYS; key++ {
		v, ok := cache.Get(key)
		assert.False(t, ok)
		assert.Nil(t, v)
	}

}

func TestSimple(t *testing.T) {
	cache := Simple(time.Millisecond*100, time.Millisecond*10)
	start := time.Now()
	const KEYS = 10
	for key := 0; key <= KEYS; key++ {
		cache.Append(key, &TimeableTest{K: int(key), V: key, Ts: start})
	}

	for key := 0; key <= KEYS; key++ {
		v, ok := cache.Get(key)
		assert.True(t, ok)
		assert.NotNil(t, v)
	}

	time.Sleep(time.Millisecond * 200)
	for key := 0; key <= KEYS; key++ {
		v, ok := cache.Get(key)
		assert.False(t, ok)
		assert.Nil(t, v)
	}

}

func TestBench(t *testing.T) {
	maxProcs := runtime.GOMAXPROCS(0)
	fmt.Printf("Max procs now: %v\n", maxProcs)
	factory := &Cache{}
	start := time.Now().UTC()
	const KEYS int32 = 200
	var keys [KEYS]int
	var MS = 100
	var VALS_PER_MS = 100
	var n_expired int32
	var total_n_expired int32

	arr := make([]*TimeableTest, 0, VALS_PER_MS*MS)
	for i := 0; i < MS; i++ { //"milliseconds"
		for k := 0; k < VALS_PER_MS; k++ {
			key := rand.Int31n(KEYS)
			val := keys[key]
			keys[key] += 1
			arr = append(arr, &TimeableTest{K: int(key), V: val, Ts: start.Add(time.Millisecond * time.Duration(i))})
		}

	}

	cache := factory.Custom(func(key interface{}, arr []Timeable) {
		atomic.AddInt32(&n_expired, 1)
		for _, _ = range arr {
			atomic.AddInt32(&total_n_expired, 1)
		}
	},
		time.Millisecond*200,
		make(chan Timeable),
		20,
		100)

	//	fmt.Println(len(arr))
	for _, a := range arr {
		cache.Append(a.K, a)
	}
	st := time.Now()
	its := int32(1000 * 100)
	var CONCURRENCY = maxProcs * 3
	wg := sync.WaitGroup{}
	wg.Add(CONCURRENCY)
	for j := 0; j < CONCURRENCY; j++ {
		go func() {
			var i int32 = 0
			for ; i < its; i++ {
				key := int32(i) % KEYS
				val := keys[key]
				keys[key] += 1
				cache.Append(key, &TimeableTest{K: int(key), V: val, Ts: start.Add(time.Millisecond * time.Duration(i))})
			}
			wg.Done()
		}()
	}
	wg.Wait()
	passed := time.Since(st)
	fmt.Printf("passed: %v, per one iteration: %v, per second: %v, per physical core: %v",
		passed,
		passed/(time.Duration(CONCURRENCY)*time.Duration(its)),
		int((time.Duration(CONCURRENCY)*time.Duration(its))/(passed/time.Second)),
		int((time.Duration(CONCURRENCY)*time.Duration(its))/(passed/time.Second))/int(maxProcs/2),
	)

	//fmt.Printf("total: %v, nkeys: %v", total_n_expired, n_expired)
}

func BenchmarkCache1(b *testing.B) {

	maxProcs := runtime.GOMAXPROCS(0)
	fmt.Printf("Max procs now: %v\n", maxProcs)

	factory := &Cache{}
	start := time.Now().UTC()
	const KEYS int32 = 200
	var keys [KEYS]int
	var MS = 100
	var VALS_PER_MS = 100
	var n_expired int32
	var total_n_expired int32

	arr := make([]*TimeableTest, 0, VALS_PER_MS*MS)
	for i := 0; i < MS; i++ { //"milliseconds"
		for k := 0; k < VALS_PER_MS; k++ {
			key := rand.Int31n(KEYS)
			val := keys[key]
			keys[key] += 1
			arr = append(arr, &TimeableTest{K: int(key), V: val, Ts: start.Add(time.Millisecond * time.Duration(i))})
		}

	}

	cache := factory.Custom(func(key interface{}, arr []Timeable) {
		atomic.AddInt32(&n_expired, 1)
		for _, _ = range arr {
			atomic.AddInt32(&total_n_expired, 1)
		}
	},
		time.Millisecond*200,
		make(chan Timeable),
		20,
		1)

	//	fmt.Println(len(arr))
	for _, a := range arr {
		cache.Append(a.K, a)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := int32(i) % KEYS
		val := keys[key]
		keys[key] += 1
		cache.Append(key, &TimeableTest{K: int(key), V: val, Ts: start.Add(time.Millisecond * time.Duration(i))})
	}
	//fmt.Printf("total: %v, nkeys: %v", total_n_expired, n_expired)
}
