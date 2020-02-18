package collector

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
	config "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ClusterLabs/ha_cluster_exporter/internal/clock"
)

func TestMetricFactory(t *testing.T) {
	SUT := NewDefaultCollector("test")
	SUT.SetDescriptor("test_metric", "", nil)

	metric := SUT.MakeGaugeMetric("test_metric", 1)

	assert.Equal(t, SUT.GetDescriptor("test_metric"), metric.Desc())
}

func TestMetricFactoryWithTimestamp(t *testing.T) {
	config.Set("enable-timestamps", true)
	defer config.Set("enable-timestamps", false)

	SUT := NewDefaultCollector("test")
	SUT.Clock = &clock.StoppedClock{}
	SUT.SetDescriptor("test_metric", "", nil)

	metric := SUT.MakeGaugeMetric("test_metric", 1)
	metricDto := &dto.Metric{}
	err := metric.Write(metricDto)

	assert.Nil(t, err, "Unexpected error")

	assert.Equal(t, int64(clock.TEST_TIMESTAMP), *metricDto.TimestampMs)
}
