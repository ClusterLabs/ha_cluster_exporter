package collector

import (
	"github.com/ClusterLabs/ha_cluster_exporter/internal/clock"
	"github.com/go-kit/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"os"
)

const NAMESPACE = "ha_cluster"

type SubsystemCollector interface {
	GetSubsystem() string
}

type DefaultCollector struct {
	subsystem   string
	descriptors map[string]*prometheus.Desc
	Clock       clock.Clock
	timestamps  bool
	Logger      log.Logger
}

func NewDefaultCollector(subsystem string, timestamps bool, logger log.Logger) DefaultCollector {
	return DefaultCollector{
		subsystem,
		make(map[string]*prometheus.Desc),
		&clock.SystemClock{},
		timestamps,
		logger,
	}
}

func (c *DefaultCollector) GetDescriptor(name string) *prometheus.Desc {
	desc, ok := c.descriptors[name]
	if !ok {
		// we hard panic on this because it's most certainly a coding error
		panic(errors.Errorf("undeclared metric '%s'", name))
	}
	return desc
}

// Convenience wrapper around prometheus.NewDesc constructor.
// Stores a metric descriptor with a fully qualified name like `NAMESPACE_subsystem_name`.
// `name` is the last and most relevant part of the metrics Full Qualified Name;
// `help` is the message displayed in the HELP line
// `variableLabels` is a list of labels to declare. Use `nil` to declare no labels.
func (c *DefaultCollector) SetDescriptor(name, help string, variableLabels []string) {
	c.descriptors[name] = prometheus.NewDesc(prometheus.BuildFQName(NAMESPACE, c.subsystem, name), help, variableLabels, nil)
}

func (c *DefaultCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descriptor := range c.descriptors {
		ch <- descriptor
	}
}

func (c *DefaultCollector) GetSubsystem() string {
	return c.subsystem
}

func (c *DefaultCollector) MakeGaugeMetric(name string, value float64, labelValues ...string) prometheus.Metric {
	return c.makeMetric(name, value, prometheus.GaugeValue, labelValues...)
}

func (c *DefaultCollector) MakeCounterMetric(name string, value float64, labelValues ...string) prometheus.Metric {
	return c.makeMetric(name, value, prometheus.CounterValue, labelValues...)
}

func (c *DefaultCollector) makeMetric(name string, value float64, valueType prometheus.ValueType, labelValues ...string) prometheus.Metric {
	desc := c.GetDescriptor(name)
	metric := prometheus.MustNewConstMetric(desc, valueType, value, labelValues...)
	if c.timestamps == true {
		metric = prometheus.NewMetricWithTimestamp(c.Clock.Now(), metric)
	}
	return metric
}

// check that all the given paths exist and are executable files
func CheckExecutables(paths ...string) error {
	for _, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil || os.IsNotExist(err) {
			return errors.Errorf("'%s' does not exist", path)
		}
		if fileInfo.IsDir() {
			return errors.Errorf("'%s' is a directory", path)
		}
		if (fileInfo.Mode() & 0111) == 0 {
			return errors.Errorf("'%s' is not executable", path)
		}
	}
	return nil
}
