package gounter

import (
	"math"
	"sync"
	"sync/atomic"
)

var releaseLocker = &sync.Mutex{}

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
	// fix release wrong
	releaseLocker.Lock()
	defer releaseLocker.Unlock()

	c.reset()
	maxCounterPool.Put(c)
}

// reset MaxCounter
// And releases Counter
func (c *MaxCounter) reset() {
	counter := c.counter
	ReleaseCounter(counter)

	c.counter = nil
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
func (c *MaxCounter) Set(value float64) {
	max := c.GetMax()
	if max < value {
		return
	}

	c.counter.Set(value)
}

// Get a number.
// if counter number > max, return max;
// else return true number.
func (c *MaxCounter) Get() float64 {
	val := c.counter.Get()
	max := c.GetMax()

	if val > max {
		return max
	}

	return val
}

// Real get Counter Real().
func (c *MaxCounter) Real() float64 {
	return c.counter.Real()
}

// Label returns CounterWithMax.
func (c *MaxCounter) Label() CounterType {
	return CounterWithMax
}

// Reset reset MaxCounter.
func (c *MaxCounter) Reset() {
	c.reset()
}

// Add is same as Counter.Add().
func (c *MaxCounter) Add(delta float64) {
	if c.isDone() {
		return
	}

	max := c.GetMax()
	realNum := c.Real()

	if realNum >= max {
		c.setDone()
		return
	}

	c.counter.Add(delta)
}

// Sub is same as Counter.Sub().
func (c *MaxCounter) Sub(delta float64) {
	c.Add(delta * -1)
}

// Inc is same as Counter.Inc().
func (c *MaxCounter) Inc() {
	c.Add(1)
}

// Dec is same as Counter.Dec().
func (c *MaxCounter) Dec() {
	c.Add(-1)
}

// CopyTo copies number to dst
func (c *MaxCounter) CopyTo(dst *MaxCounter) (ok bool, err error) {
	if c == dst {
		return
	}

	if c.Label() != dst.Label() {
		return
	}

	for {
		oldMaxBits := atomic.LoadUint64(&dst.maxBits)
		bits := atomic.LoadUint64(&c.maxBits)

		oldCounter := dst.counter
		counter := c.counter

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
