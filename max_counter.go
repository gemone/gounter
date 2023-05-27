package gounter

import (
	"math"
	"sync"
	"sync/atomic"
)

// MaxCounter has a max number for counter.
// When counter to max number, it will stop and reject all other actions.
type MaxCounter struct {
	noCopy noCopy

	counter *Counter

	done    uint32
	maxBits uint64
}

// maxCounterPool
var maxCounterPool = &sync.Pool{
	New: func() any {
		return &MaxCounter{}
	},
}

// AcquireMaxCounter acquire a MaxCounter from pool.
func AcquireMaxCounter(max float64) *MaxCounter {
	maxCounter := maxCounterPool.Get().(*MaxCounter)
	maxCounter.SetMax(max)
	maxCounter.counter = AcquireCounter()

	return maxCounter
}

// ReleaseMaxCounter releases MaxCounter.
func ReleaseMaxCounter(c *MaxCounter) {
	if c == nil {
		return
	}

	c.reset()
	ReleaseCounter(c.counter)
	c.counter = nil
	maxCounterPool.Put(c)
}

// reset MaxCounter
// And releases Counter
func (c *MaxCounter) reset() {
	if c.counter != nil {
		c.counter.Reset()
	} else {
		c.counter = AcquireCounter()
	}
	atomic.StoreUint64(&c.maxBits, 0)
	atomic.StoreUint32(&c.done, 0)
}

// isDone say now is max?
func (c *MaxCounter) isDone() bool {
	done := atomic.LoadUint32(&c.done)

	return done != 0
}

// setDone set add done, now is max.
func (c *MaxCounter) setDone() {
	atomic.StoreUint32(&c.done, 1)
}

// setUnDone set done value = 0.
func (c *MaxCounter) setUnDone() {
	atomic.StoreUint32(&c.done, 0)
}

// Can use add?
func (c *MaxCounter) Can() bool {
	return !c.isDone()
}

// GetMax gets a max number.
func (c *MaxCounter) GetMax() float64 {
	bits := atomic.LoadUint64(&c.maxBits)
	max := math.Float64frombits(bits)

	return max
}

// SetMax set a max number.
func (c *MaxCounter) SetMax(max float64) {
	for {
		oldBits := atomic.LoadUint64(&c.maxBits)
		newBits := math.Float64bits(max)

		if atomic.CompareAndSwapUint64(&c.maxBits, oldBits, newBits) {
			return
		}
	}
}

// Set sets the value of the counter to the given value,
// if it is less than or equal to the maximum value
func (c *MaxCounter) Set(value float64) bool {
	max := c.GetMax()
	if max < value {
		return false
	}

	return c.counter.Set(value)
}

// Get a number.
func (c *MaxCounter) Get() float64 {
	return c.counter.Get()
}

// Real get Counter Real().
func (c *MaxCounter) Real() float64 {
	return c.counter.Real()
}

// Reset reset MaxCounter.
func (c *MaxCounter) Reset() {
	c.reset()
}

// Add is same as Counter.Add().
func (c *MaxCounter) Add(delta float64) bool {
	if c.isDone() && delta >= 0 {
		return false
	}

	if c.isDone() && delta < 0 {
		c.setUnDone()
	}

	max := c.GetMax()
	realNum := c.Real()

	if realNum >= max && delta > 0 {
		c.setDone()
		return false
	}

	if realNum <= 0 && delta < 0 {
		return false
	}

	return c.counter.Add(delta)
}

// Sub is same as Counter.Sub().
func (c *MaxCounter) Sub(delta float64) bool {
	return c.Add(delta * -1)
}

// Inc is same as Counter.Inc().
func (c *MaxCounter) Inc() bool {
	return c.Add(1)
}

// Dec is same as Counter.Dec().
func (c *MaxCounter) Dec() bool {
	return c.Add(-1)
}

// CopyTo copies number to dst.
func (c *MaxCounter) CopyTo(d interface{}) (ok bool, err error) {
	dst, can := d.(*MaxCounter)
	if !can {
		err = ErrDifferentCounterType
		return
	}

	if c == dst {
		err = ErrSameCounterPointer
		return
	}

	for {
		oldMaxBits := atomic.LoadUint64(&dst.maxBits)
		bits := atomic.LoadUint64(&c.maxBits)

		oldCounter := dst.counter
		counter := c.counter

		if counter == nil {
			counter = AcquireCounter()
		}

		if oldCounter == nil {
			oldCounter = AcquireCounter()
		}

		ok, err = counter.CopyTo(oldCounter)
		if !ok {
			return
		}

		dst.counter = oldCounter

		if atomic.CompareAndSwapUint64(&dst.maxBits, oldMaxBits, bits) {
			ok = true
			return
		}
	}
}
