package clock

import "time"

type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}
