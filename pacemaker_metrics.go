package main

import (
	"encoding/xml"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// *** crm_mon XML unserialization structures
type pacemakerStatus struct {
	Version string  `xml:"version,attr"`
	Summary summary `xml:"summary"`
	Nodes   struct {
		Node []node `xml:"node"`
	} `xml:"nodes"`
	NodeHistory struct {
		Node []struct {
			Name            string `xml:"name,attr"`
			ResourceHistory []struct {
				Name               string `xml:"id,attr"`
				MigrationThreshold int    `xml:"migration-threshold,attr"`
				FailCount          int    `xml:"fail-count,attr"`
			} `xml:"resource_history"`
		} `xml:"node"`
	} `xml:"node_history"`
}

type summary struct {
	Nodes struct {
		Number int `xml:"number,attr"`
	} `xml:"nodes_configured"`
	LastChange struct {
		Time string `xml:"time,attr"`
	} `xml:"last_change"`
	Resources struct {
		Number   int `xml:"number,attr"`
		Disabled int `xml:"disabled,attr"`
		Blocked  int `xml:"blocked,attr"`
	} `xml:"resources_configured"`
	ClusterOptions struct {
		StonithEnabled bool `xml:"stonith-enabled,attr"`
	} `xml:"cluster_options"`
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

// Pacemaker CIB is queried from cibadmin and unmarshaled from XML
type CIB struct {
	Configuration struct {
		Constraints struct{
			RscLocations []struct {
				Id  string `xml:"id,attr"`
				Resource string `xml:"rsc,attr"`
				Score string `xml:"score,attr"`
				Node string `xml:"node,attr"`
			} `xml:"rsc_location"`
		} `xml:"constraints"`
	} `xml:"configuration"`
}

var (
	pacemakerMetrics = metricDescriptors{
		// the map key will function as an identifier of the metric throughout the rest of the code;
		// it is arbitrary, but by convention we use the actual metric name
		"nodes":               NewMetricDesc("pacemaker", "nodes", "The nodes in the cluster; one line per name, per status", []string{"node", "type", "status"}),
		"nodes_total":         NewMetricDesc("pacemaker", "nodes_total", "Total number of nodes in the cluster", nil),
		"resources":           NewMetricDesc("pacemaker", "resources", "The resources in the cluster; one line per id, per status", []string{"node", "resource", "role", "managed", "status"}),
		"resources_total":     NewMetricDesc("pacemaker", "resources_total", "Total number of resources in the cluster", nil),
		"stonith_enabled":     NewMetricDesc("pacemaker", "stonith_enabled", "Whether or not stonith is enabled", nil),
		"fail_count":          NewMetricDesc("pacemaker", "fail_count", "The Fail count number per node and resource id", []string{"node", "resource"}),
		"migration_threshold": NewMetricDesc("pacemaker", "migration_threshold", "The migration_threshold number per node and resource id", []string{"node", "resource"}),
		"config_last_change":  NewMetricDesc("pacemaker", "config_last_change", "The timestamp of the last change of the cluster configuration", nil),
		"constraints":         NewMetricDesc("pacemaker", "constraints", "Indicate if a constraints of specific type is present per ID and per resource", []string{"constraint", "resource", "node"}),
	}
)

func NewPacemakerCollector(crmMonPath string, cibAdminPath string) (*pacemakerCollector, error) {
	err := CheckExecutables(crmMonPath, cibAdminPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize Pacemaker collector")
	}

	return &pacemakerCollector{
		DefaultCollector{
			metrics: pacemakerMetrics,
		},
		crmMonPath,
		cibAdminPath,
	}, nil
}

type pacemakerCollector struct {
	DefaultCollector
	crmMonPath   string
	cibAdminPath string
}

func (c *pacemakerCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Infoln("Collecting pacemaker metrics...")

	pacemakerStatus, err := c.getPacemakerStatus()
	if err != nil {
		log.Warnln(err)
		return
	}

	cib, err := c.getCIB()
	if err != nil {
		log.Warnln(err)
		return
	}

	var stonithEnabled float64
	if pacemakerStatus.Summary.ClusterOptions.StonithEnabled {
		stonithEnabled = 1
	}

	ch <- c.makeGaugeMetric("nodes_total", float64(pacemakerStatus.Summary.Nodes.Number))
	ch <- c.makeGaugeMetric("resources_total", float64(pacemakerStatus.Summary.Resources.Number))
	ch <- c.makeGaugeMetric("stonith_enabled", stonithEnabled)

	c.recordNodeMetrics(pacemakerStatus, ch)
	c.recordFailCountMetrics(pacemakerStatus, ch)
	c.recordMigrationThresholdMetrics(pacemakerStatus, ch)
	c.recordResourceAgentsChanges(pacemakerStatus, ch)
	c.recordMigrationConstraintsMetrics(cib, ch)
}

func (c *pacemakerCollector) getPacemakerStatus() (pacemakerStatus, error) {
	var pacemakerStatus pacemakerStatus
	pacemakerStatusXML, err := exec.Command(c.crmMonPath, "-X", "--group-by-node", "--inactive").Output()
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
				ch <- c.makeGaugeMetric("nodes", float64(1), node.Name, nodeType, nodeStatus)
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
				ch <- c.makeGaugeMetric(
					"resources",
					float64(1),
					node.Name,
					resource.ID,
					strings.ToLower(resource.Role),
					strconv.FormatBool(resource.Managed),
					resourceStatus)
			}
		}
	}
}

func (c *pacemakerCollector) recordFailCountMetrics(pacemakerStatus pacemakerStatus, ch chan<- prometheus.Metric) {
	for _, node := range pacemakerStatus.NodeHistory.Node {
		for _, resHistory := range node.ResourceHistory {
			failCount := float64(resHistory.FailCount)

			// if value is 1000000 this is a special value in pacemaker which is infinity fail count
			if resHistory.FailCount >= 1000000 {
				failCount = math.Inf(1)
			}

			ch <- c.makeGaugeMetric("fail_count", failCount, node.Name, resHistory.Name)

		}
	}
}

func (c *pacemakerCollector) recordResourceAgentsChanges(pacemakerStatus pacemakerStatus, ch chan<- prometheus.Metric) {
	t, err := time.Parse(time.ANSIC, pacemakerStatus.Summary.LastChange.Time)
	if err != nil {
		log.Warnln(err)
		return
	}
	// we record the timestamp of the last change as a float counter metric, which is in turn timestamped with the time it was checked
	ch <- prometheus.NewMetricWithTimestamp(clock.Now(), prometheus.MustNewConstMetric(c.metrics["config_last_change"], prometheus.CounterValue, float64(t.Unix())))
}

func (c *pacemakerCollector) recordMigrationThresholdMetrics(pacemakerStatus pacemakerStatus, ch chan<- prometheus.Metric) {
	for _, node := range pacemakerStatus.NodeHistory.Node {
		for _, resHistory := range node.ResourceHistory {
			ch <- c.makeGaugeMetric("migration_threshold", float64(resHistory.MigrationThreshold), node.Name, resHistory.Name)
		}
	}
}

func (c *pacemakerCollector) getCIB() (CIB, error) {
	var cib CIB
	cibXML, err := exec.Command(c.cibAdminPath, "--query", "--local").Output()
	if err != nil {
		return cib, errors.Wrap(err, "error while executing cibadmin")
	}

	err = xml.Unmarshal(cibXML, &cib)
	if err != nil {
		return cib, errors.Wrap(err, "could not parse cibadmin status from XML")
	}

	return cib, nil
}

func (c *pacemakerCollector) recordMigrationConstraintsMetrics(cib CIB, ch chan<- prometheus.Metric) {
	for _, constraint := range cib.Configuration.Constraints.RscLocations {
		var constraintScore float64
		switch constraint.Score {
		case "INFINITY":
			constraintScore = math.Inf(1)
		case "-INFINITY":
			constraintScore = math.Inf(-1)
		default:
			s, _ := strconv.Atoi(constraint.Score)
			constraintScore = float64(s)
		}

		ch <- c.makeGaugeMetric("constraints", constraintScore, constraint.Id, constraint.Resource, constraint.Node)
	}
}
