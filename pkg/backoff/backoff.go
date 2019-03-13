package backoff

import (
	"math/rand"
	"time"
)

// BackOff contains parameters applied to a backoff function
type BackOff struct {
	attempts int
	// Duration is multiplied by Factor each after each attempt
	Factor float64
	// this is the initial Duration
	Duration time.Duration
	// Duration of each attempt can not be greater than MaxDuration before applying Jitter
	MaxDuration time.Duration
	// Amount of jitter applied after each iteration. This can be use for randomizing backoff duration
	JitterFactor float64
}

// NextDuration returns duration for next attempt
func (b *BackOff) NextDuration() time.Duration {
	b.attempts++

	duration := b.Duration

	// calculate duration for next attempt
	if b.Factor != 0 {
		b.Duration = time.Duration(float64(b.Duration) * b.Factor)
		if b.MaxDuration > 0 && b.Duration > b.MaxDuration {
			b.Duration = b.MaxDuration
		}
	}

	if b.JitterFactor > 0 {
		duration = b.Jitter(duration)
	}

	return duration
}

// Jitter returns a duration between initial and (initial + b.JitterFactor*initial)
func (b *BackOff) Jitter(initial time.Duration) time.Duration {
	factor := b.JitterFactor
	if factor <= 0 {
		factor = 1
	}

	return initial + time.Duration(rand.Float64()*factor*float64(initial))
}

// Attempts returns number of attempts tried
func (b *BackOff) Attempts() int {
	return b.attempts
}
