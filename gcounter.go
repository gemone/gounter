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

	Add(float64)
	Sub(float64)
	Inc()
	Dec()

	CopyTo() bool
}

type GounterWithLable interface {
	Label() CounterType

	Get(string) Gounter
	Real(string) Gounter

	Reset()
	ResetLabel(string)

	Add(string, float64) Gounter
	Sub(string, float64) Gounter
	Inc(string) Gounter
	Dec(string) Gounter

	CopyTo() bool
}
