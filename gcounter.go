package gounter

type CounterType byte

const (
	// CounterNormal Normal Counter
	CounterNormal CounterType = iota // CounterWithMax Max Counter
	CounterWithMax
	// CounterWithLabel Label Counter
	CounterWithLabel
	// CounterWithLabelAndMax Label Max Counter
	CounterWithLabelAndMax
)

type Gounter interface {
	Get() float64
	Label() CounterType

	Reset()

	Set(float64) bool
	Add(float64) bool
	Sub(float64) bool
	Inc() bool
	Dec() bool

	CopyTo(Gounter) (bool, error)
}

// LabelGounter gounter with label
type LabelGounter interface {
	Get(string) (float64, Gounter)

	Reset()
	ResetLabel(string)

	Set(string, float64) (bool, Gounter)
	Add(string, float64) (bool, Gounter)
	Sub(string, float64) (bool, Gounter)
	Inc(string) (bool, Gounter)
	Dec(string) (bool, Gounter)

	// Based on the map feature,
	// replication should not be accepted. (CopyTo)
}
