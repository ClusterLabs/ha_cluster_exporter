package corosync

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ClusterLabs/ha_cluster_exporter/internal/collector"
)

const subsystem = "corosync"

func NewCollector(cfgToolPath string, quorumToolPath string, timeout time.Duration, logger *slog.Logger) (*corosyncCollector, error) {
	err := collector.CheckExecutables(cfgToolPath, quorumToolPath)
	if err != nil {
		logger.Warn("could not initialize 'corosync' collector (missing executables), but continuing", "err", err)
	}

	c := &corosyncCollector{
		collector.NewDefaultCollector(subsystem, logger),
		cfgToolPath,
		quorumToolPath,
		timeout,
		NewParser(),
	}
	c.SetDescriptor("quorate", "Whether or not the cluster is quorate", nil)
	c.SetDescriptor("rings", "The status of each Corosync ring; 1 means healthy, 0 means faulty.", []string{"ring_id", "node_id", "number", "address"})
	c.SetDescriptor("ring_errors", "The total number of faulty corosync rings", nil)
	c.SetDescriptor("member_votes", "How many votes each member node has contributed with to the current quorum", []string{"node_id", "node", "local"})
	c.SetDescriptor("quorum_votes", "Cluster quorum votes; one line per type", []string{"type"})

	return c, nil
}

type corosyncCollector struct {
	collector.DefaultCollector
	cfgToolPath    string
	quorumToolPath string
	timeout        time.Duration
	parser         Parser
}

func (c *corosyncCollector) CollectWithError(ch chan<- prometheus.Metric) error {
	c.Logger.Debug("Collecting corosync metrics...")

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// We suppress the exec errors because if any interface is faulty the tools will exit with code 1, but we still want to parse the output.
	cfgToolOutput, _ := exec.CommandContext(ctx, c.cfgToolPath, "-s").Output()
	quorumToolOutput, _ := exec.CommandContext(ctx, c.quorumToolPath, "-p").Output()

	status, err := c.parser.Parse(cfgToolOutput, quorumToolOutput)
	if err != nil {
		return fmt.Errorf("corosync parser error: %w", err)
	}

	c.collectRings(status, ch)
	c.collectRingErrors(status, ch)
	c.collectQuorate(status, ch)
	c.collectQuorumVotes(status, ch)
	c.collectMemberVotes(status, ch)

	return nil
}

func (c *corosyncCollector) Collect(ch chan<- prometheus.Metric) {
	c.Logger.Debug("Collecting corosync metrics...")

	err := c.CollectWithError(ch)
	if err != nil {
		c.Logger.Warn(c.GetSubsystem()+" collector scrape failed", "err", err)
	}
}

func (c *corosyncCollector) collectQuorumVotes(status *Status, ch chan<- prometheus.Metric) {
	ch <- c.MakeGaugeMetric("quorum_votes", float64(status.QuorumVotes.ExpectedVotes), "expected_votes")
	ch <- c.MakeGaugeMetric("quorum_votes", float64(status.QuorumVotes.HighestExpected), "highest_expected")
	ch <- c.MakeGaugeMetric("quorum_votes", float64(status.QuorumVotes.TotalVotes), "total_votes")
	ch <- c.MakeGaugeMetric("quorum_votes", float64(status.QuorumVotes.Quorum), "quorum")
}

func (c *corosyncCollector) collectQuorate(status *Status, ch chan<- prometheus.Metric) {
	var quorate float64
	if status.Quorate {
		quorate = 1
	}
	ch <- c.MakeGaugeMetric("quorate", quorate)
}

func (c *corosyncCollector) collectRingErrors(status *Status, ch chan<- prometheus.Metric) {
	var numErrors float64
	for _, ring := range status.Rings {
		if ring.Faulty {
			numErrors += 1
		}
	}
	ch <- c.MakeGaugeMetric("ring_errors", numErrors)
}

func (c *corosyncCollector) collectRings(status *Status, ch chan<- prometheus.Metric) {
	for _, ring := range status.Rings {
		var healthy float64 = 1
		if ring.Faulty {
			healthy = 0
		}
		ch <- c.MakeGaugeMetric("rings", healthy, status.RingId, status.NodeId, ring.Number, ring.Address)
	}
}

func (c *corosyncCollector) collectMemberVotes(status *Status, ch chan<- prometheus.Metric) {
	for _, member := range status.Members {
		local := "false"
		if member.Local {
			local = "true"
		}
		ch <- c.MakeGaugeMetric("member_votes", float64(member.Votes), member.Id, member.Name, local)
	}
}
