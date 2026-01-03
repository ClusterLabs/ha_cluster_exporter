package sbd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ClusterLabs/ha_cluster_exporter/internal/collector"
)

const subsystem = "sbd"

const SBD_STATUS_UNHEALTHY = "unhealthy"
const SBD_STATUS_HEALTHY = "healthy"

// NewCollector create a new sbd collector
func NewCollector(sbdPath string, sbdConfigPath string, timeout time.Duration, logger *slog.Logger) (*sbdCollector, error) {
	err := checkArguments(sbdPath, sbdConfigPath)
	if err != nil {
		logger.Warn("could not initialize 'sbd' collector (missing executables or config), but continuing", "err", err)
	}

	c := &sbdCollector{
		collector.NewDefaultCollector(subsystem, logger),
		sbdPath,
		sbdConfigPath,
		timeout,
	}

	c.SetDescriptor("devices", "SBD devices; one line per device", []string{"device", "status"})
	c.SetDescriptor("timeouts", "SBD timeouts for each device and type", []string{"device", "type"})

	return c, nil
}

func checkArguments(sbdPath string, sbdConfigPath string) error {
	if err := collector.CheckExecutables(sbdPath); err != nil {
		return err
	}
	if _, err := os.Stat(sbdConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("'%s' does not exist", sbdConfigPath)
	}
	return nil
}

type sbdCollector struct {
	collector.DefaultCollector
	sbdPath       string
	sbdConfigPath string
	timeout       time.Duration
}

func (c *sbdCollector) CollectWithError(ch chan<- prometheus.Metric) error {
	c.Logger.Debug("Collecting pacemaker metrics...")

	sbdConfiguration, err := readSdbFile(c.sbdConfigPath)
	if err != nil {
		return err
	}

	sbdDevices := getSbdDevices(sbdConfiguration)

	sbdStatuses := c.getSbdDeviceStatuses(sbdDevices)
	for sbdDev, sbdStatus := range sbdStatuses {
		ch <- c.MakeGaugeMetric("devices", 1, sbdDev, sbdStatus)
	}

	sbdWatchdogs, sbdMsgWaits := c.getSbdTimeouts(sbdDevices)
	for sbdDev, sbdWatchdog := range sbdWatchdogs {
		ch <- c.MakeGaugeMetric("timeouts", sbdWatchdog, sbdDev, "watchdog")
	}

	for sbdDev, sbdMsgWait := range sbdMsgWaits {
		ch <- c.MakeGaugeMetric("timeouts", sbdMsgWait, sbdDev, "msgwait")
	}

	return nil
}

func (c *sbdCollector) Collect(ch chan<- prometheus.Metric) {
	c.Logger.Debug("Collecting pacemaker metrics...")

	err := c.CollectWithError(ch)
	if err != nil {
		c.Logger.Warn(c.GetSubsystem()+" collector scrape failed", "err", err)
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
	regex := regexp.MustCompile(`(?m)^\s*SBD_DEVICE="?((?:[\w-/]+;?\s?)+)"?\s*$`)
	sbdDevicesLine := regex.FindStringSubmatch(string(sbdConfigRaw))

	// if SBD_DEVICE line could not be found, return 0 devices
	if sbdDevicesLine == nil {
		return nil
	}

	// split the first capture group, e.g. `/dev/foo;/dev/bar`; the 0th element is always the whole line
	sbdDevices := strings.Split(strings.TrimRight(sbdDevicesLine[1], ";"), ";")
	for i, _ := range sbdDevices {
		sbdDevices[i] = strings.TrimSpace(sbdDevices[i])
	}

	return sbdDevices
}

// this function takes a list of sbd devices and returns
// a map of SBD device names with 1 if healthy, 0 if not
func (c *sbdCollector) getSbdDeviceStatuses(sbdDevices []string) map[string]string {
	sbdStatuses := make(map[string]string)
	for _, sbdDev := range sbdDevices {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
		_, err := exec.CommandContext(ctx, c.sbdPath, "-d", sbdDev, "dump").Output()

		// in case of error the device is not healthy
		if err != nil {
			sbdStatuses[sbdDev] = SBD_STATUS_UNHEALTHY
		} else {
			sbdStatuses[sbdDev] = SBD_STATUS_HEALTHY
		}
	}

	return sbdStatuses
}

// for each sbd device, extract the watchdog and msgwait timeout via regex
func (c *sbdCollector) getSbdTimeouts(sbdDevices []string) (map[string]float64, map[string]float64) {
	sbdWatchdogs := make(map[string]float64)
	sbdMsgWaits := make(map[string]float64)
	for _, sbdDev := range sbdDevices {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
		sbdDump, _ := exec.CommandContext(ctx, c.sbdPath, "-d", sbdDev, "dump").Output()

		regexW := regexp.MustCompile(`Timeout \(msgwait\)  *: \d+`)
		regex := regexp.MustCompile(`Timeout \(watchdog\)  *: \d+`)

		msgWaitLine := regexW.FindStringSubmatch(string(sbdDump))
		watchdogLine := regex.FindStringSubmatch(string(sbdDump))

		if watchdogLine == nil || msgWaitLine == nil {
			continue
		}

		// get the timeout from the line
		regexNumber := regexp.MustCompile(`\d+`)
		watchdogTimeout := regexNumber.FindString(string(watchdogLine[0]))
		msgWaitTimeout := regexNumber.FindString(string(msgWaitLine[0]))

		// map the timeout to the device
		if s, err := strconv.ParseFloat(watchdogTimeout, 64); err == nil {
			sbdWatchdogs[sbdDev] = s
		}

		// map the timeout to the device
		if s, err := strconv.ParseFloat(msgWaitTimeout, 64); err == nil {
			sbdMsgWaits[sbdDev] = s
		}

	}
	return sbdWatchdogs, sbdMsgWaits
}
