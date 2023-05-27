package gounter

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMaxCounterReleaseNil(t *testing.T) {
	ReleaseMaxCounter(nil)
}

func TestMaxCounterSetMax(t *testing.T) {
	t.Parallel()

	testMaxCounterSetMax(t)

	testGo(t, testMaxCounterSetMax, 10)
	testGo(t, testMaxCounterSetMax, 100)
	testGo(t, testMaxCounterSetMax, 1000)
	testGo(t, testMaxCounterSetMax, 10000)
}

func TestMaxCounterGetReal(t *testing.T) {
	t.Parallel()

	c := AcquireMaxCounter(50)
	defer ReleaseMaxCounter(c)

	c.Add(50)

	// add 1
	c.Inc()

	// but should 50
	realNum := c.Real()

	if realNum != 50 {
		t.Logf("realNum should 50, but %0f", realNum)
	}
}

func TestMaxCounterChange(t *testing.T) {
	t.Parallel()

	testMaxCounterCount(t)

	testGo(t, testMaxCounterCount, 10)
	testGo(t, testMaxCounterCount, 100)
	testGo(t, testMaxCounterCount, 1000)
	testGo(t, testMaxCounterCount, 10000)
}

func TestMaxCounterSetAndGet(t *testing.T) {
	t.Parallel()

	testMaxCounterSetAndGet(t)

	testGo(t, testMaxCounterSetAndGet, 10)
	testGo(t, testMaxCounterSetAndGet, 100)
	testGo(t, testMaxCounterSetAndGet, 1000)
	testGo(t, testMaxCounterSetAndGet, 10000)
}

func TestMaxCounterAddAndSub(t *testing.T) {
	t.Parallel()

	c := AcquireMaxCounter(50)
	defer ReleaseMaxCounter(c)

	for i := 0; i < 100; i += 1 {
		ok := c.Add(1)
		if i >= 50 && ok {
			t.Error("should false, but true")

			// check done
			can := c.Can()
			if can {
				t.Fatal("can not add")
			}
		}
	}

	v := c.Get()
	if v != 50 {
		t.Fatalf("should %d, but %f", 50, v)
	}

	for i := 0; i < 100; i += 1 {
		ok := c.Sub(1)
		if i >= 50 && ok {
			t.Error("should false, but true.")
		}
	}

	v = c.Get()
	if v != 0 {
		t.Fatalf("should %d, but %f", 0, v)
	}
}

func TestMaxCounter_CopyTo(t *testing.T) {
	t.Parallel()

	testMaxCounterCopyTo(t)

	testGo(t, testMaxCounterCopyTo, 10)
	testGo(t, testMaxCounterCopyTo, 100)
	testGo(t, testMaxCounterCopyTo, 1000)
	testGo(t, testMaxCounterCopyTo, 10000)
}

func TestMaxCounter_Reset(t *testing.T) {
	t.Parallel()

	c := AcquireMaxCounter(50)
	defer ReleaseMaxCounter(c)
	for i := 0; i < 50; i++ {
		c.Inc()
	}

	v := c.Get()
	if v != 50 {
		t.Fatalf("should be %d, but %f", 50, v)
	}

	c.Reset()

	v = c.Get()
	if v != 0 {
		t.Fatalf("should be %d, but %f", 0, v)
	}
}

// testMaxCounterSetAndGet
func testMaxCounterSetAndGet(t *testing.T) {
	c := AcquireMaxCounter(100)
	defer ReleaseMaxCounter(c)

	c.Set(100)
	v := c.Get()
	if v != 100 {
		t.Fatalf("should be %d, but %f", 100, v)
	}

	ok := c.Set(101)
	if ok {
		t.Fatal("should be true, but false")
	}
}

// testMaxCounterSetMax set max counter
func testMaxCounterSetMax(t *testing.T) {
	c := AcquireMaxCounter(0)
	defer ReleaseMaxCounter(c)

	// zero
	c.SetMax(0)

	if c.GetMax() != 0 {
		t.Fatal("error in SetMax")
	}

	// other
	max := rand.Float64()
	c.SetMax(max)

	if c.GetMax() != max {
		t.Fatalf("error: max should %f, but %f", max, c.GetMax())
	}
}

func testMaxCounterCount(t *testing.T) {
	goCounter := 50
	ch := make(chan struct{}, goCounter)

	c := AcquireMaxCounter(float64(goCounter))
	defer ReleaseMaxCounter(c)

	var num uint32
	var num2 uint32

	// Add
	for i := 0; i < goCounter; i++ {
		go func() {
			c.Inc()
			atomic.AddUint32(&num, 1)

			if c.Can() {
				atomic.AddUint32(&num2, 1)
			}
			ch <- struct{}{}
		}()
	}

	for i := 0; i < goCounter; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	if num != uint32(goCounter) {
		t.Fatalf("counter num should %d, but %d", goCounter, num)
	}

	if num2 != uint32(goCounter) {
		t.Fatalf("counter num should 50, but %d", num2)
	}

	realNum := c.Real()

	if realNum != float64(goCounter) {
		t.Fatalf("realNum should 50, but %0.f", realNum)
	}

	// Sub
	wg := sync.WaitGroup{}
	for i := 0; i < goCounter; i++ {
		wg.Add(1)
		go func() {
			c.Dec()
			wg.Done()
		}()
	}

	wg.Wait()

	v := c.Get()
	if v != 0 {
		t.Fatalf("should be %d, but %f", 0, v)
	}
}

func testMaxCounterCopyTo(t *testing.T) {
	for i := 0; i < 10; i++ {
		c1 := AcquireMaxCounter(50)
		c2 := AcquireMaxCounter(50)

		// c1 to c1
		ok, err := c1.CopyTo(c1)
		if ok {
			t.Fatal("same counter should err, but not!")
		}
		if err != ErrSameCounterPointer {
			t.Fatalf("same counter should err, but %s", err)
		}
		// c2 to c2
		ok, err = c2.CopyTo(c2)
		if ok {
			t.Fatal("same counter should err, but not!")
		}
		if err != ErrSameCounterPointer {
			t.Fatalf("same counter should err, but %s", err.Error())
		}

		// copy to
		ok, err = c1.CopyTo(c2)
		if !ok {
			t.Fatalf("counter should be copied, but not, %v", err)
		}

		//	copy to other
		ok, err = c1.CopyTo(1)
		if ok || err == nil {
			t.Fatal("should err, but not")
		}

		// counter nil
		c1.counter = nil
		ok, err = c1.CopyTo(c2)
		if !ok {
			t.Fatalf("should ok, but not, %v", err)
		}

		c2.counter = nil
		ok, err = c1.CopyTo(c2)
		if !ok {
			t.Fatal("should ok, but not")
		}

		// copy same, should err
		c1Counter := c1.counter
		c1.counter = c2.counter
		ok, _ = c1.CopyTo(c2)
		if ok {
			t.Fatal("should err, but not")
		}
		// fix error
		c1.counter = c1Counter

		ReleaseMaxCounter(c2)
		ReleaseMaxCounter(c1)
	}
}
