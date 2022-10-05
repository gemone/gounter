package gounter

import (
	"math"
	"sync"
	"sync/atomic"
)

// Counter supports increasing and decreasing counter internal value.
// Only responsible for counting, without any additional content.
// Implementation reference sync.WaitGroup .
//
// Decrement to a negative number is allowed,
// but once it is less than 0, `Get()` will get 0.
// However, it is still a negative number.
// You can continue to `Add()` or `Sub()`.
//
// Copying is prohibited. Please acquire new object.
type Counter struct {
	noCopy noCopy

	bits uint64
}

// counterPool is a pool for counter
var counterPool = &sync.Pool{
	New: func() any {
		return &Counter{}
	},
}

// AcquireCounter return a
func AcquireCounter() *Counter {
	return counterPool.New().(*Counter)
}

func ReleaseCounter(c *Counter) {
	c.reset()
	counterPool.Put(c)
}

func (c *Counter) reset() {
	atomic.StoreUint64(&c.bits, 0)
}

func (c *Counter) Get() float64 {
	bits := atomic.LoadUint64(&c.bits)
	val := math.Float64frombits(bits)
	if val < 0 {
		return 0
	}

	return val
}

func (c *Counter) Inc() {
	c.Add(1)
}

func (c *Counter) Dec() {
	c.Add(-1)
}

func (c *Counter) Add(delta float64) {
	for {
		oldBits := atomic.LoadUint64(&c.bits)
		newVal := math.Float64frombits(oldBits) + delta
		newBits := math.Float64bits(newVal)
		if atomic.CompareAndSwapUint64(&c.bits, oldBits, newBits) {
			return
		}
	}
}

func (c *Counter) Sub(delta float64) {
	c.Add(delta * -1)
}

func (c *Counter) CopyTo(dst *Counter) {
	// fix c to c can not return
	if c == dst {
		return
	}

	for {
		oldVal1 := atomic.LoadUint64(&dst.bits)

		val1 := atomic.LoadUint64(&c.bits)

		if atomic.CompareAndSwapUint64(&dst.bits, oldVal1, val1) {
			return
		}
	}
}
