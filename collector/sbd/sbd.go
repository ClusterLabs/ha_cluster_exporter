package sbd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/ClusterLabs/ha_cluster_exporter/collector"
)

const subsystem = "sbd"

const SBD_STATUS_UNHEALTHY = "unhealthy"
const SBD_STATUS_HEALTHY = "healthy"

// NewCollector create a new sbd collector
func NewCollector(sbdPath string, sbdConfigPath string) (*sbdCollector, error) {
	err := checkArguments(sbdPath, sbdConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not initialize '%s' collector", subsystem)
	}

	c := &sbdCollector{
		collector.NewDefaultCollector("sbd"),
		sbdPath,
		sbdConfigPath,
	}

	c.SetDescriptor("devices", "SBD devices; one line per device", []string{"device", "status"})
	c.SetDescriptor("device_timeout", "sbd timeouts ", []string{"type"})

	return c, nil
}

func checkArguments(sbdPath string, sbdConfigPath string) error {
	if err := collector.CheckExecutables(sbdPath); err != nil {
		return err
	}
	if _, err := os.Stat(sbdConfigPath); os.IsNotExist(err) {
		return errors.Errorf("'%s' does not exist", sbdConfigPath)
	}
	return nil
}

type sbdCollector struct {
	collector.DefaultCollector
	sbdPath       string
	sbdConfigPath string
}

func (c *sbdCollector) CollectWithError(ch chan<- prometheus.Metric) error {
	log.Debugln("Collecting SBD metrics...")

	sbdConfiguration, err := readSdbFile(c.sbdConfigPath)
	if err != nil {
		return err
	}

	sbdDevices := getSbdDevices(sbdConfiguration)

	sbdStatuses := c.getSbdDeviceStatuses(sbdDevices)
	for sbdDev, sbdStatus := range sbdStatuses {
		ch <- c.MakeGaugeMetric("devices", 1, sbdDev, sbdStatus)
	}

	c.recordSbdTimeout(sbdDevices, ch)

	return nil
}

func (c *sbdCollector) Collect(ch chan<- prometheus.Metric) {
	err := c.CollectWithError(ch)
	if err != nil {
		log.Warnf("'%s' collector scrape failed: %s", c.GetSubsystem(), err)
	}
}

func readSdbFile(sbdConfigPath string) ([]byte, error) {
	sbdConfFile, err := os.Open(sbdConfigPath)
	if err != nil {
		return nil, fmt.Errorf("could not open sbd config file %s", err)
	}

	defer sbdConfFile.Close()
	sbdConfigRaw, err := ioutil.ReadAll(sbdConfFile)

	if err != nil {
		return nil, fmt.Errorf("could not read sbd config file %s", err)
	}
	return sbdConfigRaw, nil
}

// retrieve a list of sbd devices from the config file contents
func getSbdDevices(sbdConfigRaw []byte) []string {
	// The following regex matches lines like SBD_DEVICE="/dev/foo" or SBD_DEVICE=/dev/foo;/dev/bar
	// It captures all the colon separated device names, without double quotes, into a capture group
	// It allows for free indentation, trailing spaces and end of lines, and it will ignore commented lines
	// Unbalanced double quotes are not checked and they will still produce a match
	// If multiple matching lines are present, only the first will be used
	// The single device name pattern is `[\w-/]+`, which is pretty relaxed
	regex := regexp.MustCompile(`(?m)^\s*SBD_DEVICE="?((?:[\w-/]+;?)+)"?\s*$`)
	sbdDevicesLine := regex.FindStringSubmatch(string(sbdConfigRaw))

	// if SBD_DEVICE line could not be found, return 0 devices
	if sbdDevicesLine == nil {
		return nil
	}

	// split the first capture group, e.g. `/dev/foo;/dev/bar`; the 0th element is always the whole line
	sbdDevices := strings.Split(sbdDevicesLine[1], ";")

	return sbdDevices
}

// this function takes a list of sbd devices and returns
// a map of SBD device names with 1 if healthy, 0 if not
func (c *sbdCollector) getSbdDeviceStatuses(sbdDevices []string) map[string]string {
	sbdStatuses := make(map[string]string)
	for _, sbdDev := range sbdDevices {
		_, err := exec.Command(c.sbdPath, "-d", sbdDev, "dump").Output()

		// in case of error the device is not healthy
		if err != nil {
			sbdStatuses[sbdDev] = SBD_STATUS_UNHEALTHY
		} else {
			sbdStatuses[sbdDev] = SBD_STATUS_HEALTHY
		}
	}

	return sbdStatuses
}

func (c *sbdCollector) recordSbdTimeout(sbdDevices []string, ch chan<- prometheus.Metric) {

	for _, sbdDev := range sbdDevices {
		output, _ := exec.Command(c.sbdPath, "-d", sbdDev, "dump").Output()
		log.Info(output)
	}

	timeout := 5
	timeoutType := "inProgress"
	ch <- c.MakeGaugeMetric("device_timeout", float64(timeout), timeoutType)
}
