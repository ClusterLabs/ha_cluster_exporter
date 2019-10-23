package main

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type StoppedClock struct{}

func (StoppedClock) Now() time.Time {
	ms := 1234 * time.Millisecond
	return time.Date(1970, 1, 1, 0, 0, 0, int(ms.Nanoseconds()), time.UTC)
	// 1234 millisecond after Unix epoch (1970-01-01 00:00:01.234 +0000 UTC)
	// this will allow us to assert that all the metrics are timestamped with "1234"
}

// borrowed from haproxy_exporter
// https://github.com/prometheus/haproxy_exporter/blob/0ddc4bc5cb4074ba95d57257f63ab82ab451a45b/haproxy_exporter_test.go
func expectMetrics(t *testing.T, c prometheus.Collector, fixture string) {
	exp, err := os.Open(path.Join("test", fixture))
	if err != nil {
		t.Fatalf("Error opening fixture file %q: %v", fixture, err)
	}
	if err := testutil.CollectAndCompare(c, exp); err != nil {
		t.Fatal("Unexpected metrics returned:", err)
	}
}
