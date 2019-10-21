package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const NAMESPACE = "ha_cluster"

type metricDescriptors map[string]*prometheus.Desc

type DefaultCollector struct {
	metrics metricDescriptors
	mutex   sync.RWMutex
}

func (c *DefaultCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *DefaultCollector) makeGaugeMetric(metricKey string, value float64, labelValues ...string) prometheus.Metric {
	desc, ok := c.metrics[metricKey]
	if !ok {
		// we hard panic on this because it's most certainly a coding error
		panic(errors.Errorf("undeclared metric '%s'", metricKey))
	}
	return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, labelValues...)
}

func NewMetricDesc(subsystem, name, help string, variableLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(NAMESPACE, subsystem, name), help, variableLabels, nil)
}

var (
	portNumber = flag.String("port", "9002", "The port number to listen on for HTTP requests.")
)

func main() {
	// read cli option and setup initial stat
	flag.Parse()

	pacemakerCollector, err := NewPacemakerCollector()
	if err != nil {
		log.Warnf("Could not initialize Pacemaker collector: %v\n", err)
	} else {
		prometheus.MustRegister(pacemakerCollector)
	}

	corosyncCollector, err := NewCorosyncCollector()
	if err != nil {
		log.Warnf("Could not initialize Corosync collector: %v\n", err)
	} else {
		prometheus.MustRegister(corosyncCollector)
	}

	sbdCollector, err := NewSbdCollector()
	if err != nil {
		log.Warnf("Could not initialize SBD collector: %v\n", err)
	} else {
		prometheus.MustRegister(sbdCollector)
	}

	drbdCollector, err := NewDrbdCollector()
	if err != nil {
		log.Warnf("Could not initialize DRBD collector: %v\n", err)
	} else {
		prometheus.MustRegister(drbdCollector)
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Infoln("Serving metrics on port", *portNumber)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *portNumber), nil))
}
