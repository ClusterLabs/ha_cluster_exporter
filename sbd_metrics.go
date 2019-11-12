package main

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
)

const SBD_STATUS_UNHEALTHY = 0
const SBD_STATUS_HEALTHY = 1

var (
	sbdMetrics = metricDescriptors{
		// the map key will function as an identifier of the metric throughout the rest of the code;
		// it is arbitrary, but by convention we use the actual metric name
		"device_status": NewMetricDesc("sbd", "device_status", "Whether or not an SBD device is healthy; one line per device", []string{"device"}),
		"devices_total": NewMetricDesc("sbd", "devices_total", "Total count of configured SBD devices", nil),
	}
)

func NewSbdCollector(sbdPath string, sbdConfigPath string) (*sbdCollector, error) {
	err := CheckExecutables(sbdPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize SBD collector")
	}

	if _, err := os.Stat(sbdConfigPath); os.IsNotExist(err) {
		return nil, errors.Errorf("could not initialize SBD collector: '%s' does not exist", sbdConfigPath)
	}

	return &sbdCollector{
		DefaultCollector{
			metrics: sbdMetrics,
		},
		sbdPath,
		sbdConfigPath,
	}, nil
}

type sbdCollector struct {
	DefaultCollector
	sbdPath       string
	sbdConfigPath string
}

func (c *sbdCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Infoln("Collecting SBD metrics...")

	sbdConfiguration, err := readSdbFile(c.sbdConfigPath)
	if err != nil {
		log.Warnln(err)
		return
	}

	sbdDevices := getSbdDevices(sbdConfiguration)
	ch <- c.makeGaugeMetric("devices_total", float64(len(sbdDevices)))

	sbdStatuses := c.getSbdDeviceStatuses(sbdDevices)
	for sbdDev, sbdStatus := range sbdStatuses {
		ch <- c.makeGaugeMetric("device_status", sbdStatus, sbdDev)
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
func (c *sbdCollector) getSbdDeviceStatuses(sbdDevices []string) map[string]float64 {
	sbdStatuses := make(map[string]float64)
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
