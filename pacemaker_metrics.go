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

var (
	pacemakerMetrics = metricsGroup{
		// the map key will function as an identifier of the metric throughout the rest of the code;
		// it is arbitrary, but by convention we use the actual metric name
		"nodes":           newMetricDesc("pacemaker", "nodes", "Describes each cluster node, with multiple lines per status", []string{"name", "type", "status"}),
		"nodes_total":     newMetricDesc("pacemaker", "nodes_total", "Total number of nodes in the cluster", nil),
		"resources":       newMetricDesc("pacemaker", "resources", "Describes each cluster resource, with multiple lines per status", []string{"node", "id", "role", "managed", "status"}),
		"resources_total": newMetricDesc("pacemaker", "resources_total", "Total number of resources in the cluster", nil),
	}

	crmMonPath = "/usr/sbin/crm_mon"
)

func NewPacemakerCollector() (*pacemakerCollector, error) {
	fileInfo, err := os.Stat(crmMonPath)
	if os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "could not find '%s'", crmMonPath)
	}
	if fileInfo.Mode()&0111 != 0 {
		return nil, errors.Wrapf(err, "'%s' is not executable", crmMonPath)
	}

	return &pacemakerCollector{metrics: pacemakerMetrics}, err
}

type pacemakerCollector struct {
	metrics metricsGroup
	mutex   sync.RWMutex
}

func (c *pacemakerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *pacemakerCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Infoln("Collecting pacemaker metrics...")

	pacemakerStatus, err := getPacemakerStatus()
	if err != nil {
		log.Warnln(err)
		return
	}

	ch <- c.makeMetric("nodes_total", prometheus.GaugeValue, float64(pacemakerStatus.Summary.Nodes.Number))
	ch <- c.makeMetric("resources_total", prometheus.GaugeValue, float64(pacemakerStatus.Summary.Resources.Number))

	c.recordNodeMetrics(pacemakerStatus, ch)
}

func (c *pacemakerCollector) makeMetric(metricKey string, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	desc, ok := c.metrics[metricKey]
	if !ok {
		// we hard panic on this because it's most certainly a coding error
		panic(errors.Errorf("undeclared metric '%s'", metricKey))
	}
	return prometheus.NewMetricWithTimestamp(time.Now(), prometheus.MustNewConstMetric(desc, valueType, value, labelValues...))
}

func getPacemakerStatus() (pacemakerStatus, error) {
	var pacemakerStatus pacemakerStatus
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
	var pacemakerStatus pacemakerStatus
	err := xml.Unmarshal(pacemakerXMLRaw, &pacemakerStatus)
	if err != nil {
		return pacemakerStatus, errors.Wrap(err, "could not parse Pacemaker status from XML")
	}
	return pacemakerStatus, nil
}

func (c *pacemakerCollector) recordNodeMetrics(pacemakerStatus pacemakerStatus, ch chan<- prometheus.Metric) {
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

		c.recordResourcesMetrics(node, ch)
	}
}

func (c *pacemakerCollector) recordResourcesMetrics(node node, ch chan<- prometheus.Metric) {
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
