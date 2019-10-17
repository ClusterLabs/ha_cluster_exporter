package main

import (
	"encoding/xml"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// this types are for reading pacemaker configuration xml when running crm_mon command
// and lookup the corrispective value
type pacemakerStatus struct {
	Version string  `xml:"version,attr"`
	Summary summary `xml:"summary"`
	Nodes   nodes   `xml:"nodes"`
}

type summary struct {
	Nodes struct {
		Number int `xml:"number,attr"`
	} `xml:"nodes_configured"`
	Resources resourcesConfigured `xml:"resources_configured"`
}

type resourcesConfigured struct {
	Number   int `xml:"number,attr"`
	Disabled int `xml:"disabled,attr"`
	Blocked  int `xml:"blocked,attr"`
}

type nodes struct {
	Node []node `xml:"node"`
}

type node struct {
	Name             string     `xml:"name,attr"`
	ID               string     `xml:"id,attr"`
	Online           bool       `xml:"online,attr"`
	Standby          bool       `xml:"standby,attr"`
	StandbyOnFail    bool       `xml:"standby_onfail,attr"`
	Maintenance      bool       `xml:"maintenance,attr"`
	Pending          bool       `xml:"pending,attr"`
	Unclean          bool       `xml:"unclean,attr"`
	Shutdown         bool       `xml:"shutdown,attr"`
	ExpectedUp       bool       `xml:"expected_up,attr"`
	DC               bool       `xml:"is_dc,attr"`
	ResourcesRunning int        `xml:"resources_running,attr"`
	Type             string     `xml:"type,attr"`
	Resources        []resource `xml:"resource"`
}

type resource struct {
	ID             string `xml:"id,attr"`
	Agent          string `xml:"resource_agent,attr"`
	Role           string `xml:"role,attr"`
	Active         bool   `xml:"active,attr"`
	Orphaned       bool   `xml:"orphaned,attr"`
	Blocked        bool   `xml:"blocked,attr"`
	Managed        bool   `xml:"managed,attr"`
	Failed         bool   `xml:"failed,attr"`
	FailureIgnored bool   `xml:"failure_ignored,attr"`
	NodesRunningOn int    `xml:"nodes_running_on,attr"`
}

func parsePacemakerStatus(pacemakerXMLRaw []byte) (pacemakerStatus, error) {
	var pacemakerStat pacemakerStatus
	err := xml.Unmarshal(pacemakerXMLRaw, &pacemakerStat)
	if err != nil {
		return pacemakerStat, errors.Wrap(err, "could not parse Pacemaker status from XML")
	}
	return pacemakerStat, nil
}

type metricsGroup map[string]*prometheus.Desc

type pacemakerCollector struct {
	metrics metricsGroup
}

func NewPacemakerCollector() *pacemakerCollector {
	return &pacemakerCollector{
		metrics: metricsGroup{
			"nodes_total": prometheus.NewDesc(prometheus.BuildFQName(NAMESPACE, "nodes", "total"), "Total number of nodes in the cluster", nil, nil),
		},
	}
}

func (c *pacemakerCollector) makeMetric(metricKey string, valueType prometheus.ValueType, value float64) (prometheus.Metric, error) {
	metric, err := prometheus.NewConstMetric(c.metrics[metricKey], valueType, value)
	if err != nil {
		return nil, errors.Wrap(err, "could not build a new metric")
	}
	return prometheus.NewMetricWithTimestamp(time.Now(), metric), nil
}

func (c *pacemakerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *pacemakerCollector) Collect(ch chan<- prometheus.Metric) {
	// get cluster status xml
	log.Infoln("Reading cluster status with crm_mon...")
	pacemakerXMLRaw, err := exec.Command("/usr/sbin/crm_mon", "-1", "--as-xml", "--group-by-node", "--inactive").Output()
	if err != nil {
		log.Warnln(err)
		return
	}

	// parse raw XML returned from crm_mon and populate structs for metrics
	data, err := parsePacemakerStatus(pacemakerXMLRaw)
	if err != nil {
		log.Warnln(err)
		return
	}

	nodesTotal, err := prometheus.NewConstMetric(c.metrics["nodes_total"], prometheus.GaugeValue, float64(data.Summary.Nodes.Number))
	nodesTotal = prometheus.NewMetricWithTimestamp(time.Now(), nodesTotal)
	if err != nil {
		log.Warnln(err)
		return
	}
	ch <- nodesTotal
}
