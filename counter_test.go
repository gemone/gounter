package gounter

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewCounter(t *testing.T) {
	t.Parallel()

	testNewCounter(t)

	testGo(t, testNewCounter, 10)
	testGo(t, testNewCounter, 100)
	testGo(t, testNewCounter, 1000)
	testGo(t, testNewCounter, 10000)
}

func TestCounterCopyTo(t *testing.T) {
	t.Parallel()

	testCopyTo(t)

	testGo(t, testCopyTo, 10)
	testGo(t, testCopyTo, 100)
	testGo(t, testCopyTo, 1000)
	testGo(t, testCopyTo, 10000)
}

func TestCounterChange(t *testing.T) {
	t.Parallel()

	testCounterInc(t)
	testCounterIncAndDec(t)
	testCounterDecZero(t)

	testGo(t, testCounterInc, 10)
	testGo(t, testCounterIncAndDec, 10)
	testGo(t, testCounterDecZero, 10)
}

func testNewCounter(t *testing.T) {
	for i := 0; i < 10; i++ {
		c := AcquireCounter()

		if c.Get() != 0 {
			t.Fatalf("error: bits=%d", c.bits)
		}

		// change val1
		atomic.AddUint64(&c.bits, 2)

		// Test Release
		ReleaseCounter(c)
	}
}

func testCopyTo(t *testing.T) {
	for i := 0; i < 10; i++ {
		c1 := AcquireCounter()
		c2 := AcquireCounter()

		// change val1
		v1Change := rand.Uint64()

		atomic.AddUint64(&c1.bits, v1Change)

		// c1 to c1
		c1.CopyTo(c1)
		// c2 to c2
		c2.CopyTo(c2)

		// copyto
		c1.CopyTo(c2)

		if c2.bits != v1Change {
			t.Fatalf("copy error: val=%d", c2.bits)
		}

		ReleaseCounter(c2)
		ReleaseCounter(c1)
	}
}

func testCounterInc(t *testing.T) {
	ch := make(chan struct{}, 100)

	c := AcquireCounter()
	defer ReleaseCounter(c)

	for i := 0; i < 100; i++ {
		go func() {
			c.Inc()
			ch <- struct{}{}
		}()
	}

	for i := 0; i < 100; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	if c.Get() != 100 {
		t.Fatalf("inc error: should 100, but %f", c.Get())
	}
}

func testCounterIncAndDec(t *testing.T) {
	ch := make(chan struct{}, 200)

	c := AcquireCounter()
	defer ReleaseCounter(c)

	for i := 0; i < 100; i++ {
		go func() {
			c.Inc()
			ch <- struct{}{}
		}()

		go func() {
			c.Dec()
			ch <- struct{}{}
		}()
	}

	for i := 0; i < 200; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	if c.Get() != 0 {
		t.Fatalf("inc error: should 0, but %f", c.Get())
	}
}

func testCounterDecZero(t *testing.T) {
	ch := make(chan struct{}, 200)

	c := AcquireCounter()
	defer ReleaseCounter(c)

	for i := 0; i < 100; i++ {
		go func() {
			c.Sub(rand.Float64())
			ch <- struct{}{}
		}()

		go func() {
			c.Dec()
			ch <- struct{}{}
		}()
	}

	for i := 0; i < 200; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	if c.Get() != 0 {
		t.Fatalf("inc error: should 0, but %f", c.Get())
	}
}
