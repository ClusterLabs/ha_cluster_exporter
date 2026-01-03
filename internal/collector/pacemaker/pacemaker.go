package pacemaker

import (
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ClusterLabs/ha_cluster_exporter/internal/collector"
	"github.com/ClusterLabs/ha_cluster_exporter/internal/collector/pacemaker/cib"
	"github.com/ClusterLabs/ha_cluster_exporter/internal/collector/pacemaker/crmmon"

	"github.com/prometheus/client_golang/prometheus"
)

const subsystem = "pacemaker"

func NewCollector(crmMonPath string, cibAdminPath string, timeout time.Duration, logger *slog.Logger) (*pacemakerCollector, error) {
	err := collector.CheckExecutables(crmMonPath, cibAdminPath)
	if err != nil {
		logger.Warn("could not initialize 'pacemaker' collector (missing executables), but continuing", "err", err)
	}

	c := &pacemakerCollector{
		collector.NewDefaultCollector(subsystem, logger),
		crmmon.NewCrmMonParser(crmMonPath, timeout),
		cib.NewCibAdminParser(cibAdminPath, timeout),
	}
	c.SetDescriptor("nodes", "The status of each node in the cluster; 1 means the node is in that status, 0 otherwise", []string{"node", "type", "status"})
	c.SetDescriptor("node_attributes", "Metadata attributes of each node; value is always 1", []string{"node", "name", "value"})
	c.SetDescriptor("resources", "The status of each resource in the cluster; 1 means the resource is in that status, 0 otherwise", []string{"node", "resource", "role", "managed", "status", "agent", "group", "clone"})
	c.SetDescriptor("stonith_enabled", "Whether or not stonith is enabled", nil)
	c.SetDescriptor("maintenance_mode_enabled", "Whether or not cluster wide maintenance-mode is enabled", nil)
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

func (c *pacemakerCollector) CollectWithError(ch chan<- prometheus.Metric) error {
	c.Logger.Debug("Collecting pacemaker metrics...")

	crmMon, err := c.crmMonParser.Parse()
	if err != nil {
		return fmt.Errorf("crm_mon parser error: %w", err)
	}

	CIB, err := c.cibParser.Parse()
	if err != nil {
		return fmt.Errorf("cibadmin parser error: %w", err)
	}

	c.recordStonithStatus(crmMon, ch)
	c.recordMaintenanceModeStatus(crmMon, ch)
	c.recordNodes(crmMon, ch)
	c.recordNodeAttributes(crmMon, ch)
	c.recordResources(crmMon, ch)
	c.recordFailCounts(crmMon, ch)
	c.recordMigrationThresholds(crmMon, ch)
	c.recordConstraints(CIB, ch)

	err = c.recordCibLastChange(crmMon, ch)
	if err != nil {
		return fmt.Errorf("could not record CIB last change: %w", err)
	}

	return nil
}

func (c *pacemakerCollector) Collect(ch chan<- prometheus.Metric) {
	c.Logger.Debug("Collecting pacemaker metrics...")

	err := c.CollectWithError(ch)
	if err != nil {
		c.Logger.Warn(c.GetSubsystem()+" collector scrape failed", "err", err)
	}
}

func (c *pacemakerCollector) recordMaintenanceModeStatus(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	var maintenanceModeEnabled float64
	if crmMon.Summary.ClusterOptions.MaintenanceMode {
		maintenanceModeEnabled = 1
	}

	ch <- c.MakeGaugeMetric("maintenance_mode_enabled", maintenanceModeEnabled)
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
		recorded := make(map[crmmon.Resource]bool) // we need to track cloned resources to avoid duplicates
		for _, resource := range clone.Resources {
			// Avoid recording stopped cloned resources multiple times
			if recorded[resource] == true {
				continue
			}

			c.recordResource(resource, "", clone.Id, ch)

			recorded[resource] = true
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

	var nodeName string
	if resource.Node != nil {
		nodeName = resource.Node.Name
	}

	// since we have a combined cardinality of resource * status, we cycle through all the possible statuses
	// and we record a new metric if the flag for that status is on
	for resourceStatus, flag := range resourceStatuses {
		var statusValue float64
		if flag {
			statusValue = 1
		}

		labels := []string{
			nodeName,
			resource.Id,
			strings.ToLower(resource.Role),
			strconv.FormatBool(resource.Managed),
			resourceStatus,
			resource.Agent,
			group,
			clone,
		}

		ch <- c.MakeGaugeMetric("resources", statusValue, labels...)
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
	t, err := time.ParseInLocation(time.ANSIC, crmMon.Summary.LastChange.Time, time.Local)
	if err != nil {
		return fmt.Errorf("could not parse date: %w", err)
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

func (c *pacemakerCollector) recordNodeAttributes(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	for _, node := range crmMon.NodeAttributes.Nodes {
		for _, attr := range node.Attributes {
			ch <- c.MakeGaugeMetric("node_attributes", 1, node.Name, attr.Name, attr.Value)
		}
	}
}
