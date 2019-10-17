package main

import (
	"encoding/xml"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// *** crm_mon XML unserialization structures
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

// ***

var pacemakerMetrics = metricsGroup{
	"nodes":           newPacemakerMetric("nodes", "Describes each cluster node", []string{"name", "type", "status"}),
	"nodes_total":     newPacemakerMetric("nodes_total", "Total number of nodes in the cluster", nil),
	"resources":       newPacemakerMetric("resources", "Describes each cluster resource", []string{"node", "id", "role", "managed", "status"}),
	"resources_total": newPacemakerMetric("resources_total", "Total number of resources in the cluster", nil),
}

func newPacemakerMetric(name, help string, variableLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(NAMESPACE, "pacemaker", name), help, variableLabels, nil)
}

func NewPacemakerCollector(crmMonPath string) (*pacemakerCollector, error) {
	if _, err := os.Stat(crmMonPath); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "could not find crm_mon at '%s'", crmMonPath)
	}
	return &pacemakerCollector{
		metrics:    pacemakerMetrics,
		crmMonPath: crmMonPath,
	}, nil
}

type pacemakerCollector struct {
	metrics    metricsGroup
	mutex      sync.RWMutex
	crmMonPath string
}

func (c *pacemakerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *pacemakerCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Debugln("Collecting pacemaker metrics...")

	pacemakerStatus, err := fetchPacemakerStatus(c.crmMonPath)
	if err != nil {
		log.Warnln(err)
		return
	}

	ch <- c.makeMetric("nodes_total", prometheus.GaugeValue, float64(pacemakerStatus.Summary.Nodes.Number))
	ch <- c.makeMetric("resource_total", prometheus.GaugeValue, float64(pacemakerStatus.Summary.Resources.Number))

	c.recordNodesMetrics(pacemakerStatus, ch)
	c.recordResourcesMetrics(pacemakerStatus, ch)
}

func (c *pacemakerCollector) makeMetric(metricKey string, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	return prometheus.NewMetricWithTimestamp(time.Now(), prometheus.MustNewConstMetric(c.metrics[metricKey], valueType, value, labelValues...))
}

func fetchPacemakerStatus(crmMonPath string) (pacemakerStatus pacemakerStatus, err error) {
	pacemakerStatusXML, err := exec.Command(crmMonPath, "-X", "--group-by-node", "--inactive").Output()
	if err != nil {
		return pacemakerStatus, errors.Wrap(err, "error while executing crm_mon")
	}

	pacemakerStatus, err = parsePacemakerStatus(pacemakerStatusXML)
	if err != nil {
		return pacemakerStatus, errors.Wrap(err, "error while parsing crm_mon XML output")
	}

	return pacemakerStatus, nil
}

func parsePacemakerStatus(pacemakerXMLRaw []byte) (pacemakerStatus, error) {
	var pacemakerStat pacemakerStatus
	err := xml.Unmarshal(pacemakerXMLRaw, &pacemakerStat)
	if err != nil {
		return pacemakerStat, errors.Wrap(err, "could not parse Pacemaker status from XML")
	}
	return pacemakerStat, nil
}

func (c *pacemakerCollector) recordNodesMetrics(pacemakerStatus pacemakerStatus, ch chan<- prometheus.Metric) {
	for _, node := range pacemakerStatus.Nodes.Node {
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

func (c *pacemakerCollector) recordResourcesMetrics(pacemakerStatus pacemakerStatus, ch chan<- prometheus.Metric) {
	for _, node := range pacemakerStatus.Nodes.Node {
		for _, resource := range node.Resources {
			resourceStatuses := map[string]bool{
				"active":          resource.Active,
				"orphaned":        resource.Orphaned,
				"blocked":         resource.Blocked,
				"failed":          resource.Failed,
				"failure_ignored": resource.FailureIgnored,
			}
			for resourceStatus, isActive := range resourceStatuses {
				if isActive {
					ch <- c.makeMetric(
						"resources",
						prometheus.GaugeValue,
						float64(1),
						node.Name,
						strings.ToLower(resource.ID),
						strings.ToLower(resource.Role),
						strconv.FormatBool(resource.Managed),
						resourceStatus)
				}
			}
		}
	}
}
