package clock

import "time"

type StoppedClock struct{}

const TEST_TIMESTAMP = 1234

func (StoppedClock) Now() time.Time {
	ms := TEST_TIMESTAMP * time.Millisecond
	return time.Date(1970, 1, 1, 0, 0, 0, int(ms.Nanoseconds()), time.UTC)
	// 1234 millisecond after Unix epoch (1970-01-01 00:00:01.234 +0000 UTC)
	// this will allow us to use a fixed timestamped when running assertions
}

func (StoppedClock) Since(t time.Time) time.Duration {
	return TEST_TIMESTAMP * time.Millisecond
}
