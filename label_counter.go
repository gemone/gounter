package gounter

import (
	"sync"
)

// LabelCounter is a structure that combines Counter and Label relationships,
// and has the characteristics of both Map and Counter.
// It cannot be copied directly.
type LabelCounter[T Gounter] struct {
	noCopy noCopy

	value   []T
	labels  map[string]int
	entries map[int]string
	acq     func() T
	rel     func(T)
	mux     sync.RWMutex
}

// NewLabelCounter creates and returns a new LabelCounter,
// which is a generic struct that stores the mapping between values and labels.
func NewLabelCounter[T Gounter](acq func() T, rel func(T)) *LabelCounter[T] {
	return &LabelCounter[T]{
		value:   make([]T, 0),
		labels:  make(map[string]int),
		entries: make(map[int]string),
		acq:     acq,
		rel:     rel,
	}
}

// NewLabelCounterNormal returns a new LabelCounter with Counter as the underlying type.
// It uses AcquireCounter and ReleaseCounter as the acquire and release functions for Counter.
func NewLabelCounterNormal() *LabelCounter[*Counter] {
	return NewLabelCounter[*Counter](AcquireCounter, ReleaseCounter)
}

// NewLabelCounterWithMax returns a new LabelCounter with MaxCounter as the underlying type.
// It uses AcquireMaxCounter and ReleaseMaxCounter as the acquire and release functions for MaxCounter.
// The max parameter is the maximum value for the MaxCounter.
func NewLabelCounterWithMax(max float64) *LabelCounter[*MaxCounter] {
	acq := func() *MaxCounter {
		return AcquireMaxCounter(max)
	}

	return NewLabelCounter[*MaxCounter](acq, ReleaseMaxCounter)
}

// RemoveLabel removes the label and its associated counter value from the LabelCounter.
// It returns true if the label was found and removed, false otherwise.
// It also releases the counter value using the rel function.
func (counter *LabelCounter[T]) RemoveLabel(label string) (ok bool) {
	counter.mux.Lock()
	defer counter.mux.Unlock()

	index, ok := counter.labels[label]
	if !ok {
		// not found
		return true
	}

	c := counter.value[index]

	// release the counter value using the rel function
	counter.rel(c)

	lastIdx := len(counter.value) - 1

	// if the last index is not equal to the current index
	if lastIdx != index {
		var lastLabel string
		lastLabel, ok = counter.entries[lastIdx]
		if !ok {
			// not found, something is wrong
			// TODO: check not found
			return
		}

		// load the counter value of the last label from the value slice
		cc := counter.value[lastIdx]
		// replace the current index with the last index in the value slice
		counter.value[index] = cc
		// update the labels map with the new index for the last label
		counter.labels[lastLabel] = index
		// delete the last index from the entries map
		delete(counter.entries, lastIdx)
	}

	// remove label
	counter.value = counter.value[:lastIdx]
	delete(counter.labels, label)

	return true
}

// newLabel creates a new counter value using the acq function and associates it with the given label.
// It returns the counter value and its index in the value slice.
// It also stores the label and its index in the labels and entries maps.
// Just use in getLabel.
func (counter *LabelCounter[T]) newLabel(label string) (c T, idx int) {
	c = counter.acq()

	counter.value = append(counter.value, c)
	idx = len(counter.value) - 1
	counter.labels[label] = idx
	counter.entries[idx] = label

	return
}

// getLabel returns the counter value and index for the given label.
// If the label is not found in the counter.labels map,
// it calls newLabel to create a new counter value and index for it.
func (counter *LabelCounter[T]) getLabel(label string, justGet bool) (c T, idx int) {
	counter.mux.Lock()
	defer counter.mux.Unlock()

	idx, ok := counter.labels[label]
	if !ok {
		if justGet {
			idx = -1
			return
		}
		return counter.newLabel(label)
	}

	if len(counter.value) <= idx {
		return counter.newLabel(label)
	}

	return counter.value[idx], idx
}

// Get returns the value and the Gounter associated with the given label.
func (counter *LabelCounter[T]) Get(label string) (v float64, c T) {
	c, idx := counter.getLabel(label, true)

	if idx == -1 {
		v = 0
		return
	}

	return c.Get(), c
}

// Set sets the value of the Gounter associated with the given label to the given value.
func (counter *LabelCounter[T]) Set(label string, v float64) (ok bool, c T) {
	c, _ = counter.getLabel(label, false)

	ok = c.Set(v)
	return
}

// Reset resets the counter to an empty state.
func (counter *LabelCounter[T]) Reset() {
	counter.mux.Lock()
	defer counter.mux.Unlock()

	counter.value = counter.value[:0]
	counter.entries = make(map[int]string)
	counter.labels = make(map[string]int)
}

// ResetLabel resets the counter for the given label to zero.
func (counter *LabelCounter[T]) ResetLabel(label string) {
	c, idx := counter.getLabel(label, true)

	if idx == -1 {
		return
	}

	c.Reset()
}

// Add adds the given delta to the counter for the given label
// and returns the updated value and a boolean indicating success or failure.
func (counter *LabelCounter[T]) Add(label string, delta float64) (ok bool, c T) {
	c, _ = counter.getLabel(label, false)
	ok = c.Add(delta)
	return
}

// Sub subtracts the given delta from the counter for the given label
// and returns the updated value and a boolean indicating success or failure.
func (counter *LabelCounter[T]) Sub(label string, delta float64) (ok bool, c T) {
	c, idx := counter.getLabel(label, true)
	if idx == -1 {
		return
	}

	ok = c.Sub(delta)
	return
}

// Inc increments the counter for the given label by one
// and returns the updated value and a boolean indicating success or failure.
func (counter *LabelCounter[T]) Inc(label string) (ok bool, c T) {
	c, _ = counter.getLabel(label, false)
	ok = c.Inc()
	return
}

// Dec decrements the counter for the given label by one
// and returns the updated value and a boolean indicating success or failure.
func (counter *LabelCounter[T]) Dec(label string) (ok bool, c T) {
	c, idx := counter.getLabel(label, true)
	if idx == -1 {
		return
	}

	ok = c.Dec()
	return
}
