package corosync

import (
	"os/exec"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/ClusterLabs/ha_cluster_exporter/collector"
)

func NewCollector(cfgToolPath string, quorumToolPath string) (*corosyncCollector, error) {
	err := collector.CheckExecutables(cfgToolPath, quorumToolPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize Corosync collector")
	}

	c := &corosyncCollector{
		collector.NewDefaultCollector("corosync"),
		cfgToolPath,
		quorumToolPath,
		NewParser(),
	}
	c.SetDescriptor("quorate", "Whether or not the cluster is quorate", nil)
	c.SetDescriptor("rings", "The status of each Corosync ring; 1 means healthy, 0 means faulty.", []string{"id", "number", "address"})
	c.SetDescriptor("ring_errors", "The number of faulty corosync rings", nil)
	c.SetDescriptor("quorum_votes", "Cluster quorum votes; one line per type", []string{"type"})

	return c, nil
}

type corosyncCollector struct {
	collector.DefaultCollector
	cfgToolPath    string
	quorumToolPath string
	cfgToolParser Parser
}

func (c *corosyncCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting corosync metrics...")

	// We suppress the exec errors because if any interface is faulty the tools will exit with code 1, but we still want to parse the output.
	cfgToolOutput, _ := exec.Command(c.cfgToolPath, "-s").Output()
	quorumToolOutput, _ := exec.Command(c.quorumToolPath).Output()

	status, err := c.cfgToolParser.Parse(cfgToolOutput, quorumToolOutput)
	if err != nil {
		log.Warnf("Corosync Collector scrape failed: %s", err)
		return
	}

	c.collectRings(status, ch)
	c.collectRingErrors(status, ch)
	c.collectQuorate(status, ch)
	c.collectQuorumVotes(status, ch)
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
		ch <- c.MakeGaugeMetric("rings", healthy, status.RingId, ring.Number, ring.Address)
	}
}
