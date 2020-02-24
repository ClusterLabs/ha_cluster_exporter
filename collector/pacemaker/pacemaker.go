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
	c.SetDescriptor("nodes", "The nodes in the cluster; one line per name, per status", []string{"node", "type", "status"})
	c.SetDescriptor("resources", "The resources in the cluster; one line per id, per status", []string{"node", "resource", "role", "managed", "status"})
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
	log.Infoln("Collecting pacemaker metrics...")

	crmMon, err := c.crmMonParser.Parse()
	if err != nil {
		log.Warnln(err)
		return
	}

	CIB, err := c.cibParser.Parse()
	if err != nil {
		log.Warnln(err)
		return
	}

	var stonithEnabled float64
	if crmMon.Summary.ClusterOptions.StonithEnabled {
		stonithEnabled = 1
	}

	ch <- c.MakeGaugeMetric("stonith_enabled", stonithEnabled)

	c.recordNodes(crmMon, ch)
	c.recordResources(crmMon, ch)
	c.recordFailCounts(crmMon, ch)
	c.recordMigrationThresholds(crmMon, ch)
	c.recordResourceAgentsChanges(crmMon, ch)
	c.recordConstraints(CIB, ch)
}

func (c *pacemakerCollector) recordNodes(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	for _, node := range crmMon.Nodes {
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
				ch <- c.MakeGaugeMetric("nodes", float64(1), node.Name, nodeType, nodeStatus)
			}
		}

		c.recordNodeResources(node, ch)
	}
}

func (c *pacemakerCollector) recordNodeResources(node crmmon.Node, ch chan<- prometheus.Metric) {
	for _, resource := range node.Resources {
		resourceStatuses := map[string]bool{
			"active":          resource.Active,
			"orphaned":        resource.Orphaned,
			"blocked":         resource.Blocked,
			"failed":          resource.Failed,
			"failure_ignored": resource.FailureIgnored,
		}
		for resourceStatus, flag := range resourceStatuses {
			if flag {
				ch <- c.MakeGaugeMetric(
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

func (c *pacemakerCollector) recordResources(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	for _, resource := range crmMon.Resources {
		ch <- c.MakeGaugeMetric(
			"resources",
			float64(1),
			"",
			resource.ID,
			strings.ToLower(resource.Role),
			strconv.FormatBool(resource.Managed),
			"")
	}
}

func (c *pacemakerCollector) recordFailCounts(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	for _, node := range crmMon.NodeHistory.Node {
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

func (c *pacemakerCollector) recordResourceAgentsChanges(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	t, err := time.Parse(time.ANSIC, crmMon.Summary.LastChange.Time)
	if err != nil {
		log.Warnln(err)
		return
	}
	// we record the timestamp of the last change as a float counter metric
	ch <- c.MakeCounterMetric("config_last_change", float64(t.Unix()))
}

func (c *pacemakerCollector) recordMigrationThresholds(crmMon crmmon.Root, ch chan<- prometheus.Metric) {
	for _, node := range crmMon.NodeHistory.Node {
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
