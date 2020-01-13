package main

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	config "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type StoppedClock struct{}

const TEST_TIMESTAMP = 1234

func (StoppedClock) Now() time.Time {
	ms := TEST_TIMESTAMP * time.Millisecond
	return time.Date(1970, 1, 1, 0, 0, 0, int(ms.Nanoseconds()), time.UTC)
	// 1234 millisecond after Unix epoch (1970-01-01 00:00:01.234 +0000 UTC)
	// this will allow us to use a fixed timestamped when running assertions
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

func TestMetricFactory(t *testing.T) {
	SUT := &DefaultCollector{
		metrics: metricDescriptors{
			"test_metric": NewMetricDesc("test", "metric", "", nil),
		},
	}

	metric := SUT.makeGaugeMetric("test_metric", 1)

	assert.Equal(t, SUT.metrics["test_metric"], metric.Desc())
}

func TestMetricFactoryWithTimestamp(t *testing.T) {
	config.Set("timestamp", true)
	defer func() {
		config.Set("timestamp", false)
	}()

	clock = StoppedClock{}
	SUT := &DefaultCollector{
		metrics: metricDescriptors{
			"test_metric": NewMetricDesc("test", "metric", "", nil),
		},
	}

	metric := SUT.makeGaugeMetric("test_metric", 1)
	metricDto := &dto.Metric{}
	err := metric.Write(metricDto)

	assert.Nil(t, err, "Unexpected error")

	assert.Equal(t, int64(TEST_TIMESTAMP), *metricDto.TimestampMs)
}
