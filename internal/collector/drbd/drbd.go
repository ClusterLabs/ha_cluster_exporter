package drbd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ClusterLabs/ha_cluster_exporter/internal/collector"
)

const subsystem = "drbd"

// drbdStatus is for parsing relevant data we want to convert to metrics
type drbdStatus struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Devices []struct {
		Volume    int    `json:"volume"`
		Written   int    `json:"written"`
		Read      int    `json:"read"`
		AlWrites  int    `json:"al-writes"`
		BmWrites  int    `json:"bm-writes"`
		UpPending int    `json:"upper-pending"`
		LoPending int    `json:"lower-pending"`
		Quorum    bool   `json:"quorum"`
		DiskState string `json:"disk-state"`
	} `json:"devices"`
	Connections []struct {
		PeerNodeID  int    `json:"peer-node-id"`
		PeerRole    string `json:"peer-role"`
		PeerDevices []struct {
			Volume        int     `json:"volume"`
			Received      int     `json:"received"`
			Sent          int     `json:"sent"`
			Pending       int     `json:"pending"`
			Unacked       int     `json:"unacked"`
			PeerDiskState string  `json:"peer-disk-state"`
			PercentInSync float64 `json:"percent-in-sync"`
		} `json:"peer_devices"`
	} `json:"connections"`
}

func NewCollector(drbdSetupPath string, drbdSplitBrainPath string, timeout time.Duration, logger *slog.Logger) (*drbdCollector, error) {
	err := collector.CheckExecutables(drbdSetupPath)
	if err != nil {
		logger.Warn("could not initialize 'drbd' collector (missing executables), but continuing", "err", err)
	}

	c := &drbdCollector{
		collector.NewDefaultCollector(subsystem, logger),
		drbdSetupPath,
		drbdSplitBrainPath,
		timeout,
	}

	c.SetDescriptor("resources", "The DRBD resources; 1 line per name, per volume", []string{"resource", "role", "volume", "disk_state"})
	c.SetDescriptor("written", "KiB written to DRBD; 1 line per res, per volume", []string{"resource", "volume"})
	c.SetDescriptor("read", "KiB read from DRBD; 1 line per res, per volume", []string{"resource", "volume"})
	c.SetDescriptor("al_writes", "Writes to activity log; 1 line per res, per volume", []string{"resource", "volume"})
	c.SetDescriptor("bm_writes", "Writes to bitmap; 1 line per res, per volume", []string{"resource", "volume"})
	c.SetDescriptor("upper_pending", "Upper pending; 1 line per res, per volume", []string{"resource", "volume"})
	c.SetDescriptor("lower_pending", "Lower pending; 1 line per res, per volume", []string{"resource", "volume"})
	c.SetDescriptor("quorum", "Quorum status per resource and per volume", []string{"resource", "volume"})
	c.SetDescriptor("connections", "The DRBD resource connections; 1 line per per resource, per peer_node_id", []string{"resource", "peer_node_id", "peer_role", "volume", "peer_disk_state"})
	c.SetDescriptor("connections_sync", "The in sync percentage value for DRBD resource connections", []string{"resource", "peer_node_id", "volume"})
	c.SetDescriptor("connections_received", "KiB received per connection", []string{"resource", "peer_node_id", "volume"})
	c.SetDescriptor("connections_sent", "KiB sent per connection", []string{"resource", "peer_node_id", "volume"})
	c.SetDescriptor("connections_pending", "Pending value per connection", []string{"resource", "peer_node_id", "volume"})
	c.SetDescriptor("connections_unacked", "Unacked value per connection", []string{"resource", "peer_node_id", "volume"})
	c.SetDescriptor("split_brain", "Whether a split brain has been detected; 1 line per resource, per volume.", []string{"resource", "volume"})

	return c, nil
}

type drbdCollector struct {
	collector.DefaultCollector
	drbdsetupPath      string
	drbdSplitBrainPath string
	timeout            time.Duration
}

func (c *drbdCollector) CollectWithError(ch chan<- prometheus.Metric) error {
	c.Logger.Debug("Collecting DRBD metrics...")

	c.recordDrbdSplitBrainMetric(ch)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	drbdStatusRaw, err := exec.CommandContext(ctx, c.drbdsetupPath, "status", "--json").Output()
	if err != nil {
		return fmt.Errorf("drbdsetup command failed: %w", err)
	}
	// populate structs and parse relevant info we will expose via metrics
	drbdDev, err := parseDrbdStatus(drbdStatusRaw)
	if err != nil {
		return fmt.Errorf("could not parse drbdsetup status output: %w", err)
	}

	for _, resource := range drbdDev {
		for _, device := range resource.Devices {
			// the `resources` metric value is always 1, otherwise it's absent
			ch <- c.MakeGaugeMetric("resources", float64(1), resource.Name, resource.Role, strconv.Itoa(device.Volume), strings.ToLower(device.DiskState))
			ch <- c.MakeGaugeMetric("written", float64(device.Written), resource.Name, strconv.Itoa(device.Volume))
			ch <- c.MakeGaugeMetric("read", float64(device.Read), resource.Name, strconv.Itoa(device.Volume))
			ch <- c.MakeGaugeMetric("al_writes", float64(device.AlWrites), resource.Name, strconv.Itoa(device.Volume))
			ch <- c.MakeGaugeMetric("bm_writes", float64(device.BmWrites), resource.Name, strconv.Itoa(device.Volume))
			ch <- c.MakeGaugeMetric("upper_pending", float64(device.UpPending), resource.Name, strconv.Itoa(device.Volume))
			ch <- c.MakeGaugeMetric("lower_pending", float64(device.LoPending), resource.Name, strconv.Itoa(device.Volume))

			if device.Quorum == true {
				ch <- c.MakeGaugeMetric("quorum", float64(1), resource.Name, strconv.Itoa(device.Volume))
			} else {
				ch <- c.MakeGaugeMetric("quorum", float64(0), resource.Name, strconv.Itoa(device.Volume))
			}
		}
		if len(resource.Connections) == 0 {
			c.Logger.Warn("Could not retrieve connection info for resource "+resource.Name, "err", err)
			continue
		}
		// a Resource can have multiple connection with different nodes
		for _, conn := range resource.Connections {
			if len(conn.PeerDevices) == 0 {
				c.Logger.Warn("Could not retrieve any peer device info for connection "+resource.Name, "err", err)
				continue
			}
			for _, peerDev := range conn.PeerDevices {
				ch <- c.MakeGaugeMetric("connections", float64(1), resource.Name, strconv.Itoa(conn.PeerNodeID), conn.PeerRole, strconv.Itoa(peerDev.Volume), strings.ToLower(peerDev.PeerDiskState))
				ch <- c.MakeGaugeMetric("connections_sync", float64(peerDev.PercentInSync), resource.Name, strconv.Itoa(conn.PeerNodeID), strconv.Itoa(peerDev.Volume))
				ch <- c.MakeGaugeMetric("connections_received", float64(peerDev.Received), resource.Name, strconv.Itoa(conn.PeerNodeID), strconv.Itoa(peerDev.Volume))
				ch <- c.MakeGaugeMetric("connections_sent", float64(peerDev.Sent), resource.Name, strconv.Itoa(conn.PeerNodeID), strconv.Itoa(peerDev.Volume))
				ch <- c.MakeGaugeMetric("connections_pending", float64(peerDev.Pending), resource.Name, strconv.Itoa(conn.PeerNodeID), strconv.Itoa(peerDev.Volume))
				ch <- c.MakeGaugeMetric("connections_unacked", float64(peerDev.Unacked), resource.Name, strconv.Itoa(conn.PeerNodeID), strconv.Itoa(peerDev.Volume))
			}
		}
	}

	return nil
}

func (c *drbdCollector) Collect(ch chan<- prometheus.Metric) {
	c.Logger.Debug("Collecting DRBD metrics...")

	err := c.CollectWithError(ch)
	if err != nil {
		c.Logger.Warn(c.GetSubsystem()+" collector scrape failed", "err", err)
	}
}

func parseDrbdStatus(statusRaw []byte) ([]drbdStatus, error) {
	var drbdDevs []drbdStatus
	err := json.Unmarshal(statusRaw, &drbdDevs)
	if err != nil {
		return drbdDevs, err
	}
	return drbdDevs, nil
}

func (c *drbdCollector) recordDrbdSplitBrainMetric(ch chan<- prometheus.Metric) {
	// look for files created by the DRBD split brain hook
	files, _ := filepath.Glob(c.drbdSplitBrainPath + "/drbd-split-brain-detected-*")

	// prepare some pattern matching
	re := regexp.MustCompile(`drbd-split-brain-detected-(?P<resource>[\w-]+)-(?P<volume>[\w-]+)`)

	// for each of these files, we extract the name of the resource end volume from its name and record the metric
	for _, f := range files {
		// matches[0] will be the whole file name, matches[1] the resource, matches[2] the volume
		matches := re.FindStringSubmatch(f)
		if matches == nil {
			continue
		}

		ch <- c.MakeGaugeMetric("split_brain", float64(1), matches[1], matches[2])
	}
}
