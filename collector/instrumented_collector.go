package collector

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ClusterLabs/ha_cluster_exporter/internal/clock"
)

// describes a collector that can return errors, instead of the default Prometheus one which has void Collect returns
type FailureProneCollector interface {
	Collect(ch chan<- prometheus.Metric) error
	Describe(chan<- *prometheus.Desc)
	GetSubsystem() string
}

type InstrumentedCollector struct {
	collector FailureProneCollector
	Clock clock.Clock
}

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(NAMESPACE, "scrape", "collector_duration_seconds"),
		"node_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(NAMESPACE, "scrape", "collector_success"),
		"node_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

func NewInstrumentedCollector(collector FailureProneCollector) *InstrumentedCollector {
	return &InstrumentedCollector{collector, &clock.SystemClock{}}
}

func (ic *InstrumentedCollector) Collect(ch chan<- prometheus.Metric) {
	var success float64
	begin := ic.Clock.Now()
	err := ic.collector.Collect(ch)
	duration := ic.Clock.Since(begin)
	if err == nil {
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), ic.collector.GetSubsystem())
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, ic.collector.GetSubsystem())
}

func (ic *InstrumentedCollector) Describe(ch chan<- *prometheus.Desc) {
	ic.collector.Describe(ch)
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}
