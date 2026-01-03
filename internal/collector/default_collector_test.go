package collector

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricFactory(t *testing.T) {
	SUT := NewDefaultCollector("test", slog.New(slog.NewTextHandler(os.Stdout, nil)))
	SUT.SetDescriptor("test_metric", "", nil)

	metric := SUT.MakeGaugeMetric("test_metric", 1)

	assert.Equal(t, SUT.GetDescriptor("test_metric"), metric.Desc())
}
