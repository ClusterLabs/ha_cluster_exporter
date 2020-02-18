package assert

import (
	"os"
	"path"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// borrowed from haproxy_exporter
// https://github.com/prometheus/haproxy_exporter/blob/0ddc4bc5cb4074ba95d57257f63ab82ab451a45b/haproxy_exporter_test.go
func Metrics(t *testing.T, c prometheus.Collector, fixture string) {
	exp, err := os.Open(path.Join("test", fixture))
	if err != nil {
		t.Fatalf("Error opening fixture file %q: %v", fixture, err)
	}
	if err := testutil.CollectAndCompare(c, exp); err != nil {
		t.Fatal("Unexpected metrics returned:", err)
	}
}
