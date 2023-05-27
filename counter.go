package gounter

import (
	"math"
	"sync"
	"sync/atomic"
)

// Counter supports increasing and decreasing counter internal value.
// Only responsible for counting, without any additional content.
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

// counterPool is a pool for counter.
var counterPool = &sync.Pool{
	New: func() any {
		return &Counter{}
	},
}

// AcquireCounter return a Counter Pointer.
func AcquireCounter() *Counter {
	return counterPool.Get().(*Counter)
}

// ReleaseCounter releases a Counter Pointer.
func ReleaseCounter(c *Counter) {
	c.reset()
	counterPool.Put(c)
}

// reset Counter to release.
func (c *Counter) reset() {
	atomic.StoreUint64(&c.bits, 0)
}

// Get returns a number.
// When the counter value is negative, it returns 0.
func (c *Counter) Get() float64 {
	val := c.Real()
	if val < 0 {
		return 0
	}

	return val
}

// Real returns a number in counter.
func (c *Counter) Real() float64 {
	bits := atomic.LoadUint64(&c.bits)
	val := math.Float64frombits(bits)

	return val
}

// Inc increases the counter by 1.
// Counter always returns true.
func (c *Counter) Inc() bool {
	return c.Add(1)
}

// Dec decreases the counter by 1.
// Counter always returns true.
func (c *Counter) Dec() bool {
	return c.Add(-1)
}

// Set sets the value of the counter to the given value using atomic operations.
func (c *Counter) Set(value float64) bool {
	for {
		oldBits := atomic.LoadUint64(&c.bits)
		newBits := math.Float64bits(value)
		if atomic.CompareAndSwapUint64(&c.bits, oldBits, newBits) {
			return true
		}
	}
}

// Add increases the counter number.
// Decreasing use negative number.
// Counter always returns true.
func (c *Counter) Add(delta float64) bool {
	for {
		oldBits := atomic.LoadUint64(&c.bits)
		newVal := math.Float64frombits(oldBits) + delta
		newBits := math.Float64bits(newVal)
		if atomic.CompareAndSwapUint64(&c.bits, oldBits, newBits) {
			return true
		}
	}
}

// Sub decreases the counter number.
// Counter always returns true.
func (c *Counter) Sub(delta float64) bool {
	return c.Add(delta * -1)
}

// Reset resets this Counter.
func (c *Counter) Reset() {
	c.reset()
}

// CopyTo copies a Counter to other Counter.
// If same Counter, do not change everything.
func (c *Counter) CopyTo(d interface{}) (ok bool, err error) {
	dst, can := d.(*Counter)
	if !can {
		err = ErrDifferentCounterType
		return
	}

	// fix c to c can not return
	if c == dst {
		err = ErrSameCounterPointer
		return
	}

	for {
		oldVal1 := atomic.LoadUint64(&dst.bits)
		val1 := atomic.LoadUint64(&c.bits)
		if atomic.CompareAndSwapUint64(&dst.bits, oldVal1, val1) {
			ok = true
			return
		}
	}
}
