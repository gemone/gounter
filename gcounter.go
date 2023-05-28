package gounter

import "errors"

var (
	ErrSameCounterPointer   = errors.New("can not copy same counter")
	ErrDifferentCounterType = errors.New("can not copy different type counter")
)

type Gounter interface {
	Get() float64
	Reset()

	Set(float64) bool
	Add(float64) bool
	Sub(float64) bool
	Inc() bool
	Dec() bool

	CopyTo(interface{}) (bool, error)
}

// LabelGounter gounter with label
type LabelGounter[T Gounter] interface {
	Get(string) (float64, T)

	Reset()
	ResetLabel(string)
	RemoveLabel(string)

	Set(string, float64) (bool, T)
	Add(string, float64) (bool, T)
	Sub(string, float64) (bool, T)
	Inc(string) (bool, T)
	Dec(string) (bool, T)
	// Based on the map feature,
	// replication should not be accepted. (CopyTo)
}
