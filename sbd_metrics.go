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
		"device_status": NewMetricDesc("sbd", "device_status", "Health status for each SBD device; 1 is healthy, 0 is not", []string{"device_name"}),
	}

	sbdConfigPath = "/etc/sysconfig/sbd"
	sbdPath       = "/usr/sbin/sbd"
)

func NewSbdCollector() (*sbdCollector, error) {
	if _, err := os.Stat(sbdConfigPath); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "'%s' not found", sbdConfigPath)
	}

	fileInfo, err := os.Stat(sbdPath)
	if err != nil || os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "'%s' not found", sbdPath)
	}
	if (fileInfo.Mode() & 0111) == 0 {
		return nil, errors.Errorf("'%s' is not executable", sbdPath)
	}

	return &sbdCollector{
		DefaultCollector{
			metrics: sbdMetrics,
		},
	}, nil
}

type sbdCollector struct {
	DefaultCollector
}

func (c *sbdCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Infoln("Collecting SBD metrics...")

	sbdConfiguration, err := readSdbFile()
	if err != nil {
		log.Warnln(err)
		return
	}

	sbdDevices, err := getSbdDevices(sbdConfiguration)
	if err != nil {
		// most likely, the sbd_device were not set in config file
		log.Warnln(err)
		return
	}

	sbdStatuses, err := getSbdDeviceStatuses(sbdDevices)
	if err != nil {
		log.Warnln(err)
		return
	}
	for sbdDev, sbdStatus := range sbdStatuses {
		ch <- c.makeMetric("device_status", prometheus.GaugeValue, sbdStatus, sbdDev)
	}
}

func readSdbFile() ([]byte, error) {
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
func getSbdDevices(sbdConfigRaw []byte) ([]string, error) {
	// in config it can be both SBD_DEVICE="/dev/foo" or SBD_DEVICE=/dev/foo;/dev/bro
	wordOnly := regexp.MustCompile("SBD_DEVICE=\"?[a-zA-Z-/;]+\"?")
	sbdDevicesConfig := wordOnly.FindString(string(sbdConfigRaw))

	// check the case there is an sbd_config but the SBD_DEVICE is not set

	if sbdDevicesConfig == "" {
		return nil, fmt.Errorf("there are no SBD_DEVICE set in configuration file")
	}
	// remove the SBD_DEVICE
	sbdArray := strings.Split(sbdDevicesConfig, "SBD_DEVICE=")[1]
	// make a list of devices by ; seperators and remove double quotes if present
	sbdDevices := strings.Split(strings.Trim(sbdArray, "\""), ";")

	return sbdDevices, nil
}

// this function takes a list of sbd devices and returns
// a map of SBD device names with 1 if healthy, 0 if not
func getSbdDeviceStatuses(sbdDevices []string) (map[string]float64, error) {
	sbdStatuses := make(map[string]float64)
	for _, sbdDev := range sbdDevices {
		_, err := exec.Command(sbdPath, "-d", sbdDev, "dump").Output()

		// in case of error the device is not healthy
		if err != nil {
			sbdStatuses[sbdDev] = SBD_STATUS_UNHEALTHY
		} else {
			sbdStatuses[sbdDev] = SBD_STATUS_HEALTHY
		}
	}

	if len(sbdStatuses) == 0 {
		return nil, errors.New("could not retrieve SBD device statuses")
	}

	return sbdStatuses, nil
}
