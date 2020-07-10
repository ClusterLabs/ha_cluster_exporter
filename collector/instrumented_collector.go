package collector

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ClusterLabs/ha_cluster_exporter/internal/clock"
)

// describes a collector that can return errors, instead of the default Prometheus one, which has void Collect returns
type FailureProneCollector interface {
	prometheus.Collector
	CollectWithError(ch chan<- prometheus.Metric) error
	GetSubsystem() string
}

type InstrumentedCollector struct {
	collector FailureProneCollector
	Clock clock.Clock
	scrapeDurationDesc *prometheus.Desc
	scrapeSuccessDesc *prometheus.Desc
}

func NewInstrumentedCollector(collector FailureProneCollector) *InstrumentedCollector {
	return &InstrumentedCollector{
		collector,
		&clock.SystemClock{},
		prometheus.NewDesc(
			prometheus.BuildFQName(NAMESPACE, "scrape", "duration_seconds"),
			"Duration of a collector scrape.",
			nil,
			prometheus.Labels{
				"collector": collector.GetSubsystem(),
			},
		),
		prometheus.NewDesc(
			prometheus.BuildFQName(NAMESPACE, "scrape", "success"),
			"Whether a collector succeeded.",
			nil,
			prometheus.Labels{
				"collector": collector.GetSubsystem(),
			},
		),
	}
}

func (ic *InstrumentedCollector) Collect(ch chan<- prometheus.Metric) {
	var success float64
	begin := ic.Clock.Now()
	err := ic.collector.CollectWithError(ch)
	duration := ic.Clock.Since(begin)
	if err == nil {
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(ic.scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds())
	ch <- prometheus.MustNewConstMetric(ic.scrapeSuccessDesc, prometheus.GaugeValue, success)
}

func (ic *InstrumentedCollector) Describe(ch chan<- *prometheus.Desc) {
	ic.collector.Describe(ch)
	ch <- ic.scrapeDurationDesc
	ch <- ic.scrapeSuccessDesc
}
