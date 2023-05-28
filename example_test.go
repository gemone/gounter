package gounter

import (
	"fmt"
	"sync"
)

func ExampleCounter() {
	counter := AcquireCounter()
	defer ReleaseCounter(counter)

	wg := sync.WaitGroup{}

	wg.Add(20)
	for i := 0; i < 10; i++ {
		go func() {
			counter.Inc()
			wg.Done()
		}()
		go func() {
			counter.Dec()
			wg.Done()
		}()
	}

	counter.Inc()

	// Get
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			// Get and do other things...
			v := counter.Get()
			fmt.Println("value: ", v)
		}()
	}

	wg.Wait()
}

func ExampleMaxCounter() {
	counter := AcquireMaxCounter(50)
	defer ReleaseMaxCounter(counter)

	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 50; i++ {
		go func() {
			counter.Inc()
			wg.Done()
		}()
		go func() {
			counter.Dec()
			wg.Done()
		}()
	}

	// Get and do others
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			/// ...
		}()
	}

	wg.Wait()
}

func ExampleNewLabelCounter() {
	counter := NewLabelCounter[*Counter](AcquireCounter, ReleaseCounter)
	counter.Add("a", 1)
	// do others
}

func ExampleNewLabelCounterNormal() {
	counter := NewLabelCounterNormal()
	counter.Add("a", 1)
	// do others
}

func ExampleNewLabelCounterWithMax() {
	counter := NewLabelCounterWithMax(50)
	counter.Add("a", 10)
}
