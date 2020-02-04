package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func NewCorosyncCollector(cfgToolPath string, quorumToolPath string) (*corosyncCollector, error) {
	err := CheckExecutables(cfgToolPath, quorumToolPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize Corosync collector")
	}

	collector := &corosyncCollector{
		DefaultCollector{
			subsystem: "corosync",
		},
		cfgToolPath,
		quorumToolPath,
	}
	collector.setDescriptor("quorate", "Whether or not the cluster is quorate", nil)
	collector.setDescriptor("ring_errors_total", "Total number of corosync ring errors", nil)
	collector.setDescriptor("quorum_votes", "Cluster quorum votes; one line per type", []string{"type"})

	return collector, nil
}

type corosyncCollector struct {
	DefaultCollector
	cfgToolPath    string
	quorumToolPath string
}

func (c *corosyncCollector) Collect(ch chan<- prometheus.Metric) {
	log.Infoln("Collecting corosync metrics...")

	err := c.collectRingErrorsTotal(ch)
	if err != nil {
		log.Warnln(err)
	}

	quorumStatusRaw := c.getQuoromStatus()
	quorumStatus, quorate, err := parseQuoromStatus(quorumStatusRaw)
	if err != nil {
		log.Warnln(err)
		return
	}

	ch <- c.makeGaugeMetric("quorate", quorate)

	for voteType, value := range quorumStatus {
		ch <- c.makeGaugeMetric("quorum_votes", float64(value), voteType)
	}
}

func (c *corosyncCollector) collectRingErrorsTotal(ch chan<- prometheus.Metric) error {
	ringStatus := c.getCorosyncRingStatus()
	ringErrorsTotal, err := parseRingStatus(ringStatus)
	if err != nil {
		return errors.Wrap(err, "cannot parse ring status")
	}

	ch <- c.makeGaugeMetric("ring_errors_total", float64(ringErrorsTotal))

	return nil
}

func (c *corosyncCollector) getQuoromStatus() []byte {
	// We suppress the exec error because if any interface is faulty, the tool will exit with code 1.
	// If all interfaces are active, exit code will be 0.
	quorumInfoRaw, _ := exec.Command(c.quorumToolPath).Output()
	return quorumInfoRaw
}

func parseQuoromStatus(quoromStatusRaw []byte) (quorumVotes map[string]int, quorate float64, err error) {
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

	// In case of error, the binary is there but execution was erroring out, check output for quorate string.
	quorateWordPresent := quoratePresent.FindString(string(quoromRaw))

	// check the case there is an sbd_config but the SBD_DEVICE is not set

	if quorateWordPresent == "" {
		return nil, quorate, fmt.Errorf("cannot parse quorum status")
	}

	quorateRaw := wordOnly.FindString(strings.SplitAfterN(quoromRaw, "Quorate:", 2)[1])
	quorateString := strings.ToLower(quorateRaw)

	if quorateString == "yes" {
		quorate = 1
	}

	expVotes, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Expected votes:", 2)[1]))
	highVotes, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Highest expected:", 2)[1]))
	totalVotes, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Total votes:", 2)[1]))
	quorum, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Quorum:", 2)[1]))

	quorumVotes = map[string]int{
		"expected_votes":   expVotes,
		"highest_expected": highVotes,
		"total_votes":      totalVotes,
		"quorum":           quorum,
	}

	if len(quorumVotes) == 0 {
		return quorumVotes, quorate, fmt.Errorf("could not retrieve any quorum information")
	}

	return quorumVotes, quorate, nil
}

// get status ring and return it as bytes
// this function can return also just an malformed output in case of error, we don't check.
// It is the parser that will check the status
func (c *corosyncCollector) getCorosyncRingStatus() []byte {
	// We suppress the exec error because if any interface is faulty, the tool will exit with code 1.
	// If all interfaces are active/without error, exit code will be 0.
	ringStatusRaw, _ := exec.Command(c.cfgToolPath, "-s").Output()
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
			return 0, fmt.Errorf("corosync-cfgtool returned unexpected output: %s", statusRaw)
		}

		return 0, nil
	}

	// there is a ringError
	return ringErrorsTotal, nil
}
