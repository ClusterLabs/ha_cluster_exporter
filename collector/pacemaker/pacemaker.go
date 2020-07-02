package pacemaker

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ClusterLabs/ha_cluster_exporter/collector"
	"github.com/ClusterLabs/ha_cluster_exporter/collector/pacemaker/cib"
	"github.com/ClusterLabs/ha_cluster_exporter/collector/pacemaker/crmmon"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func NewCollector(crmMonPath string, cibAdminPath string) (*pacemakerCollector, error) {
	err := collector.CheckExecutables(crmMonPath, cibAdminPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize Pacemaker collector")
	}

	c := &pacemakerCollector{
		collector.NewDefaultCollector("pacemaker"),
		crmmon.NewCrmMonParser(crmMonPath),
		cib.NewCibAdminParser(cibAdminPath),
	}
	c.SetDescriptor("nodes", "The status of each node in the cluster; 1 means the node is in that status, 0 otherwise", []string{"node", "type", "status"})
	c.SetDescriptor("resources", "The status of each resource in the cluster; 1 means the resource is in that status, 0 otherwise", []string{"node", "resource", "role", "managed", "status", "agent", "group", "clone"})
	c.SetDescriptor("stonith_enabled", "Whether or not stonith is enabled", nil)
	c.SetDescriptor("fail_count", "The Fail count number per node and resource id", []string{"node", "resource"})
	c.SetDescriptor("migration_threshold", "The migration_threshold number per node and resource id", []string{"node", "resource"})
	c.SetDescriptor("config_last_change", "The timestamp of the last change of the cluster configuration", nil)
	c.SetDescriptor("location_constraints", "Resource location constraints. The value indicates the score.", []string{"constraint", "node", "resource", "role"})

	return c, nil
}

type pacemakerCollector struct {
	collector.DefaultCollector
	crmMonParser crmmon.Parser
	cibParser    cib.Parser
}

func (c *pacemakerCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting pacemaker metrics...")

	crmMon, err := c.crmMonParser.Parse()
	if err != nil {
		log.Warnf("Pacemaker Collector scrape failed: %s", err)
		return
	}

	CIB, err := c.cibParser.Parse()
	if err != nil {
		log.Warnf("Pacemaker Collector scrape failed: %s", err)
		return
	}

	c.recordStonithStatus(crmMon, ch)
	c.recordNodes(crmMon, ch)
	c.recordResources(crmMon, ch)
	c.recordFailCounts(crmMon, ch)
	c.recordMigrationThresholds(crmMon, ch)
	c.recordConstraints(CIB, ch)

	err = c.recordCibLastChange(crmMon, ch)
	if err != nil {
		log.Warnf("Pacemaker Collector scrape failed: %s", err)
	}
}

func (c *pacemakerCollector) recordStonithStatus(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	var stonithEnabled float64
	if crmMon.Summary.ClusterOptions.StonithEnabled {
		stonithEnabled = 1
	}

	ch <- c.MakeGaugeMetric("stonith_enabled", stonithEnabled)
}

func (c *pacemakerCollector) recordNodes(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	for _, node := range crmMon.Nodes {

		// this is a map of boolean flags for each possible status of the node
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

		// since we have a combined cardinality of node * status, we cycle through all the possible statuses
		// and we record a metric for each one
		for nodeStatus, flag := range nodeStatuses {
			var statusValue float64
			if flag {
				statusValue = 1
			}
			ch <- c.MakeGaugeMetric("nodes", statusValue, node.Name, node.Type, nodeStatus)
		}
	}
}

func (c *pacemakerCollector) recordResources(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	for _, resource := range crmMon.Resources {
		c.recordResource(resource, "", "", ch)
	}
	for _, clone := range crmMon.Clones {
		for _, resource := range clone.Resources {
			c.recordResource(resource, "", clone.Id, ch)
		}
	}
	for _, group := range crmMon.Groups {
		for _, resource := range group.Resources {
			c.recordResource(resource, group.Id, "", ch)
		}
	}
}

func (c *pacemakerCollector) recordResource(resource crmmon.Resource, group string, clone string, ch chan<- prometheus.Metric) {

	// this is a map of boolean flags for each possible status of the resource
	resourceStatuses := map[string]bool{
		"active":          resource.Active,
		"orphaned":        resource.Orphaned,
		"blocked":         resource.Blocked,
		"failed":          resource.Failed,
		"failure_ignored": resource.FailureIgnored,
	}

	// since we have a combined cardinality of resource * status, we cycle through all the possible statuses
	// and we record a new metric if the flag for that status is on
	for resourceStatus, flag := range resourceStatuses {
		var statusValue float64
		if flag {
			statusValue = 1
		}
		var nodeName string
		if resource.Node != nil {
			nodeName = resource.Node.Name
		}
		ch <- c.MakeGaugeMetric(
			"resources",
			statusValue,
			nodeName,
			resource.Id,
			strings.ToLower(resource.Role),
			strconv.FormatBool(resource.Managed),
			resourceStatus,
			resource.Agent,
			group,
			clone)
	}
}

func (c *pacemakerCollector) recordFailCounts(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	for _, node := range crmMon.NodeHistory.Nodes {
		for _, resHistory := range node.ResourceHistory {
			failCount := float64(resHistory.FailCount)

			// if value is 1000000 this is a special value in pacemaker which is infinity fail count
			if resHistory.FailCount >= 1000000 {
				failCount = math.Inf(1)
			}

			ch <- c.MakeGaugeMetric("fail_count", failCount, node.Name, resHistory.Name)

		}
	}
}

func (c *pacemakerCollector) recordCibLastChange(crmMon crmmon.Root, ch chan<- prometheus.Metric) error {
	t, err := time.Parse(time.ANSIC, crmMon.Summary.LastChange.Time)
	if err != nil {
		return errors.Wrap(err, "could not parse date")
	}
	// we record the timestamp of the last change as a float counter metric
	ch <- c.MakeCounterMetric("config_last_change", float64(t.Unix()))

	return nil
}

func (c *pacemakerCollector) recordMigrationThresholds(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	for _, node := range crmMon.NodeHistory.Nodes {
		for _, resHistory := range node.ResourceHistory {
			ch <- c.MakeGaugeMetric("migration_threshold", float64(resHistory.MigrationThreshold), node.Name, resHistory.Name)
		}
	}
}

func (c *pacemakerCollector) recordConstraints(CIB cib.Root, ch chan<- prometheus.Metric) {
	for _, constraint := range CIB.Configuration.Constraints.RscLocations {
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

		ch <- c.MakeGaugeMetric("location_constraints", constraintScore, constraint.Id, constraint.Node, constraint.Resource, strings.ToLower(constraint.Role))
	}
}
