package clock

import "time"

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

func (SystemClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}
