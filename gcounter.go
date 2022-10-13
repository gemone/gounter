package gounter

type CounterType byte

const (
	// Normal Counter
	CounterNormal CounterType = iota
	// Max Counter
	CounterWithMax
	// Label Counter
	CounterWithLabel
	// Label Max Counter
	CounterWithLabelAndMax
)

type Gounter interface {
	Get() float64
	Real() float64
	Label() CounterType

	Reset()

	Add(float64) bool
	Sub(float64) bool
	Inc() bool
	Dec() bool

	// Can returns will Gounter continue work.
	Can() bool

	CopyTo(Gounter) (bool, error)
}

type GounterWithLable interface {
	Label() CounterType

	Get(string) (float64, Gounter)
	Real(string) (float64, Gounter)

	Reset()
	ResetLabel(string)

	Add(string, float64) (bool, Gounter)
	Sub(string, float64) (bool, Gounter)
	Inc(string) (bool, Gounter)
	Dec(string) (bool, Gounter)

	Can() bool
	CanLabel(string) bool

	// Based on the map feature,
	// replication should not be accepted. (CopyTo)
}
