package email

import "time"

// SystemClock uses the Go runtime time source.
type SystemClock struct{}

// NewSystemClock constructs a SystemClock.
func NewSystemClock() Clock {
	return SystemClock{}
}

// Now returns the current UTC time.
func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
