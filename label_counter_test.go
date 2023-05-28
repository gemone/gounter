package gounter

import (
	"sync"
	"testing"
)

func testGenerateLabels() []string {
	labels := make([]string, 26*2)
	for i := 0; i < 26; i++ {
		labels[i] = string(rune('a' + i))
	}
	for i := 0; i < 26; i++ {
		labels[26+i] = string(rune('A' + i))
	}

	return labels
}

func TestLabelCounterWithNormal(t *testing.T) {
	t.Parallel()

	// generate random labels, use a-z A-z
	labels := testGenerateLabels()

	c := NewLabelCounterNormal()
	testLabelCounterIncDec(labels, 10000, c, t)
}

func TestLabelCounterWithMax(t *testing.T) {
	t.Parallel()

	labels := testGenerateLabels()

	c := NewLabelCounterWithMax(10000)
	testLabelCounterIncDec(labels, 10000, c, t)

	c1 := NewLabelCounterWithMax(1000)
	testLabelCounterIncDec(labels, 1000, c1, t)

	c2 := NewLabelCounterWithMax(100)
	testLabelCounterIncDec(labels, 100, c2, t)
}

var c = NewLabelCounterNormal()

func TestLabelCounterChange(t *testing.T) {
	t.Parallel()

	c.Set("set", 10)
	v, _ := c.Get("set")
	if v != 10 {
		t.Errorf("wrong result, expect %d, got %f", 10, v)
	}

	c.Add("set", 10)
	v, _ = c.Get("set")
	if v != 20 {
		t.Errorf("wrong result, expect %d, got %f", 20, v)
	}

	c.Sub("set", 20)
	v, _ = c.Get("set")
	if v != 0 {
		t.Errorf("wrong result, expect %d, got %f", 0, v)
	}
}

func TestLabelCounter_ChangeLabel(t *testing.T) {
	t.Parallel()

	label := "remove_label"

	c.Set(label, 10)
	if v, _ := c.Get(label); v != 10 {
		t.Errorf("wrong result, expect %d, got %f", 10, v)
	}

	c.ResetLabel(label)
	if v, _ := c.Get(label); v != 0 {
		t.Errorf("wrong result, expect %d, got %f", 0, v)
	}

	c.Add(label, 20)
	if v, _ := c.Get(label); v != 20 {
		t.Errorf("wrong result, expect %d, got %f", 20, v)
	}

	c.RemoveLabel(label)
	if v, _ := c.Get(label); v != 0 {
		t.Errorf("wrong result, expect %d, got %f", 0, v)
	}

	// multi label
	label1 := "remove_label1"
	label2 := "remove_label2"

	c.Set(label1, 10)
	c.Set(label2, 20)
	c.RemoveLabel(label1)
	c.RemoveLabel(label2)
}

func TestLabelCounter_Reset(t *testing.T) {
	t.Parallel()

	c := NewLabelCounterWithMax(20)

	c.Set("reset", 10)
	c.Set("test", 10)

	if len(c.value) != 2 {
		t.Errorf("wrong result, expect %d, got %d", 2, len(c.value))
	}

	c.Reset()

	if len(c.value) != 0 {
		t.Errorf("wrong result, expect %d, got %d", 0, len(c.value))
	}
}

// testLabelCounterIncDec tests LabelCounter.
func testLabelCounterIncDec[T Gounter](labels []string, count int, c *LabelCounter[T], t *testing.T) {
	wg := sync.WaitGroup{}

	// each label add
	for _, label := range labels {
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(ll string) {
				c.Inc(ll)
				wg.Done()
			}(label)
		}
	}

	wg.Wait()

	// assert each label add
	for _, label := range labels {
		v, _ := c.Get(label)
		if v != float64(count) {
			t.Errorf("label %s, wrong result, expect %d, got %f", label, 10, v)
		}
	}

	// decrement each label
	for _, label := range labels {
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(ll string) {
				c.Dec(ll)
				wg.Done()
			}(label)
		}
	}

	wg.Wait()

	// assert each label decrement
	for _, label := range labels {
		v, _ := c.Get(label)
		if v != 0 {
			t.Errorf("label %s, wrong result, expect %d, got %f", label, 0, v)
		}
	}
}
