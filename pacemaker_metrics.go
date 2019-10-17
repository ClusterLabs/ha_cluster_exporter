package main

import (
	"encoding/xml"
	"os/exec"
	"sync"
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

type pacemakerCollector struct {
	metrics metricsGroup
	mutex sync.RWMutex
}

var pacemakerMetrics = metricsGroup {
	"nodes_total": prometheus.NewDesc(prometheus.BuildFQName(NAMESPACE, "nodes", "total"), "Total number of nodes in the cluster", nil, nil),
	"nodes":       prometheus.NewDesc(prometheus.BuildFQName(NAMESPACE, "", "nodes"), "Describes each cluster node", []string{"name", "type", "status"}, nil),
}

func parsePacemakerStatus(pacemakerXMLRaw []byte) (pacemakerStatus, error) {
	var pacemakerStat pacemakerStatus
	err := xml.Unmarshal(pacemakerXMLRaw, &pacemakerStat)
	if err != nil {
		return pacemakerStat, errors.Wrap(err, "could not parse Pacemaker status from XML")
	}
	return pacemakerStat, nil
}

func NewPacemakerCollector() *pacemakerCollector {
	return &pacemakerCollector{metrics: pacemakerMetrics}
}

func (c *pacemakerCollector) makeMetric(metricKey string, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	return prometheus.NewMetricWithTimestamp(time.Now(), prometheus.MustNewConstMetric(c.metrics[metricKey], valueType, value, labelValues...))
}

func (c *pacemakerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *pacemakerCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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

	ch <- c.makeMetric("nodes_total", prometheus.GaugeValue, float64(data.Summary.Nodes.Number))

	c.recordClusterNodes(data, ch)
}

func (c *pacemakerCollector) recordClusterNodes(data pacemakerStatus, ch chan<- prometheus.Metric) {
	for _, node := range data.Nodes.Node {
		nodeStatuses := map[string]bool{
			"online":         node.Online,
			"standby":        node.Standby,
			"standby_onfail": node.StandbyOnFail,
			"maintenance":    node.Maintenance,
			"pending":        node.Pending,
			"unclean":        node.Unclean,
			"shutdown":       node.Shutdown,
			"expected_up":    node.ExpectedUp,
			"dc":             node.DC,
		}

		var nodeType string
		switch node.Type {
		case "member", "ping", "remote":
			nodeType = node.Type
			break
		default:
			nodeType = "unknown"
		}

		for nodeStatus, isActive := range nodeStatuses {
			if isActive {
				ch <- c.makeMetric("nodes", prometheus.GaugeValue, float64(1), node.Name, nodeType, nodeStatus)
			}
		}
	}
}
