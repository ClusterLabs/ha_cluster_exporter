package collector

import (
	"log/slog"

	"github.com/ClusterLabs/ha_cluster_exporter/internal/clock"
	"github.com/prometheus/client_golang/prometheus"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod -package mock_collector -destination ../test/mock_collector/instrumented_collector.go github.com/ClusterLabs/ha_cluster_exporter/internal/collector InstrumentableCollector

// describes a collector that can return errors from collection cycles,
// instead of the default Prometheus one, which has void Collect returns
type InstrumentableCollector interface {
	prometheus.Collector
	SubsystemCollector
	CollectWithError(ch chan<- prometheus.Metric) error
}

type InstrumentedCollector struct {
	collector          InstrumentableCollector
	Clock              clock.Clock
	scrapeDurationDesc *prometheus.Desc
	scrapeSuccessDesc  *prometheus.Desc
	logger             *slog.Logger
}

func NewInstrumentedCollector(collector InstrumentableCollector, logger *slog.Logger) *InstrumentedCollector {
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
		logger,
	}
}

func (ic *InstrumentedCollector) Collect(ch chan<- prometheus.Metric) {
	var success float64
	begin := ic.Clock.Now()
	err := ic.collector.CollectWithError(ch)
	duration := ic.Clock.Since(begin)
	if err == nil {
		success = 1
	} else {
		ic.logger.Warn(ic.collector.GetSubsystem()+" collector scrape failed", "err", err)
	}
	ch <- prometheus.MustNewConstMetric(ic.scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds())
	ch <- prometheus.MustNewConstMetric(ic.scrapeSuccessDesc, prometheus.GaugeValue, success)
}

func (ic *InstrumentedCollector) Describe(ch chan<- *prometheus.Desc) {
	ic.collector.Describe(ch)
	ch <- ic.scrapeDurationDesc
	ch <- ic.scrapeSuccessDesc
}

func (ic *InstrumentedCollector) GetSubsystem() string {
	return ic.collector.GetSubsystem()
}
