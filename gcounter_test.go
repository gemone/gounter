package gounter

import (
	"testing"
	"time"
)

func testGo(t *testing.T, f func(*testing.T), count int) {
	ch := make(chan struct{}, count)

	for i := 0; i < count; i++ {
		go func() {
			f(t)
			ch <- struct{}{}
		}()
	}

	for i := 0; i < count; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}
}
