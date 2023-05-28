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
