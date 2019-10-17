package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	corosyncMetrics = metricsGroup{
		// the map key will function as an identifier of the metric throughout the rest of the code;
		// it is arbitrary, but by convention we use the actual metric name
		"quorate":           newMetricDesc("corosync", "quorate", "Whether or not the cluster is quorate", nil),
		"ring_errors_total": newMetricDesc("corosync", "ring_errors_total", "Total number of corosync ring errors", nil),
		"quorum_votes":      newMetricDesc("corosync", "quorum_votes", "Cluster quorum votes; one line per type", []string{"type"}),
	}

	corosyncTools = map[string]string{
		"quorumtool": "/usr/sbin/corosync-quorumtool",
		"cfgtool":    "/usr/sbin/corosync-cfgtool",
	}
)

func NewCorosyncCollector() (*corosyncCollector, error) {
	for _, toolPath := range corosyncTools {
		fileInfo, err := os.Stat(toolPath)
		if os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "could not find '%s'", toolPath)
		}
		if fileInfo.Mode()&0111 != 0 {
			return nil, errors.Wrapf(err, "'%s' is not executable", toolPath)
		}
	}

	return &corosyncCollector{metrics: corosyncMetrics}, nil
}

type corosyncCollector struct {
	metrics metricsGroup
	mutex   sync.RWMutex
}

func (c *corosyncCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *corosyncCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Infoln("Collecting corosync metrics...")

	err := c.collectRingErrorsTotal(ch)
	if err != nil {
		log.Warnln(err)
	}

	quorumStatusRaw := getQuoromStatus()
	quorumStatus, isQuorate, err := parseQuoromStatus(quorumStatusRaw)
	if err != nil {
		log.Warnln(err)
		return
	}

	ch <- c.makeMetric("quorate", prometheus.GaugeValue, isQuorate)

	for voteType, value := range quorumStatus {
		ch <- c.makeMetric(
			"quorum_votes",
			prometheus.GaugeValue,
			float64(value),
			voteType)
	}
}

func (c *corosyncCollector) collectRingErrorsTotal(ch chan<- prometheus.Metric) error {
	ringStatus := getCorosyncRingStatus()
	ringErrorsTotal, err := parseRingStatus(ringStatus)
	if err != nil {
		return errors.Wrap(err, "cannot parse ring status")
	}

	ch <- c.makeMetric("ring_errors_total", prometheus.GaugeValue, float64(ringErrorsTotal))

	return nil
}

func (c *corosyncCollector) makeMetric(metricKey string, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	desc, ok := c.metrics[metricKey]
	if !ok {
		// we hard panic on this because it's most certainly a coding error
		panic(errors.Errorf("undeclared metric '%s'", metricKey))
	}
	return prometheus.NewMetricWithTimestamp(time.Now(), prometheus.MustNewConstMetric(desc, valueType, value, labelValues...))
}

func getQuoromStatus() []byte {
	// We suppress the exec error because if any interface is faulty, the tool will exit with code 1.
	// If all interfaces are active, exit code will be 0.
	quorumInfoRaw, _ := exec.Command(corosyncTools["quorumtool"]).Output()
	return quorumInfoRaw
}

func parseQuoromStatus(quoromStatusRaw []byte) (quorumStatus map[string]int, isQuorate float64, err error) {
	quoromRaw := string(quoromStatusRaw)
	// Quorate:          Yes

	// Votequorum information
	// ----------------------
	// Expected votes:   2
	// Highest expected: 2
	// Total votes:      2
	// Quorum:           1

	// We apply the same method for all the metrics/data:
	// first split the string for finding the word , e.g "Expected votes:", and get it via regex
	// only the number   2,
	// and convert it to integer type
	numberOnly := regexp.MustCompile("[0-9]+")
	wordOnly := regexp.MustCompile("[a-zA-Z]+")
	quoratePresent := regexp.MustCompile("Quorate:")

	// In case of error, the binary is there but execution was erroring out, check output for isQuorate string.
	quorateWordPresent := quoratePresent.FindString(string(quoromRaw))

	// check the case there is an sbd_config but the SBD_DEVICE is not set

	if quorateWordPresent == "" {
		return nil, isQuorate, fmt.Errorf("the quorum status output is not in parsable format as expected")
	}

	quorateRaw := wordOnly.FindString(strings.SplitAfterN(quoromRaw, "Quorate:", 2)[1])
	quorateString := strings.ToLower(quorateRaw)

	if quorateString == "yes" {
		isQuorate = 1
	}

	expVotes, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Expected votes:", 2)[1]))
	highVotes, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Highest expected:", 2)[1]))
	totalVotes, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Total votes:", 2)[1]))
	quorum, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Quorum:", 2)[1]))

	quorumVotes := map[string]int{
		"expectedVotes":   expVotes,
		"highestExpected": highVotes,
		"totalVotes":      totalVotes,
		"quorum":          quorum,
	}

	if len(quorumVotes) == 0 {
		return quorumVotes, isQuorate, fmt.Errorf("could not retrieve any quorum information")
	}

	return quorumVotes, isQuorate, nil
}

// get status ring and return it as bytes
// this function can return also just an malformed output in case of error, we don't check.
// It is the parser that will check the status
func getCorosyncRingStatus() []byte {
	// We suppress the exec error because if any interface is faulty, the tool will exit with code 1.
	// If all interfaces are active/without error, exit code will be 0.
	ringStatusRaw, _ := exec.Command(corosyncTools["cfgtool"], "-s").Output()
	return ringStatusRaw
}

// return the number of RingError that we will use as gauge, and error if somethings unexpected happens
func parseRingStatus(ringStatus []byte) (int, error) {
	statusRaw := string(ringStatus)
	// check if there is a ring ERROR first
	ringErrorsTotal := strings.Count(statusRaw, "FAULTY")

	// in case there is no error we need to check that the output is not
	if ringErrorsTotal == 0 {
		// if there is no RING ID word, the command corosync-cfgtool went wrong/error out
		if strings.Count(statusRaw, "RING ID") == 0 {
			return 0, fmt.Errorf("corosync-cfgtool command returned unexpected output %s", statusRaw)
		}

		return 0, nil
	}

	// there is a ringError
	return ringErrorsTotal, nil
}
