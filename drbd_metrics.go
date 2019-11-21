package main

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// drbdStatus is for parsing relevant data we want to convert to metrics
type drbdStatus struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Devices []struct {
		Volume    int    `json:"volume"`
		DiskState string `json:"disk-state"`
	} `json:"devices"`
	Connections []struct {
		PeerNodeID  int    `json:"peer-node-id"`
		PeerRole    string `json:"peer-role"`
		PeerDevices []struct {
			Volume        int     `json:"volume"`
			PeerDiskState string  `json:"peer-disk-state"`
			PercentInSync float64 `json:"percent-in-sync"`
		} `json:"peer_devices"`
	} `json:"connections"`
}

var (
	drbdMetrics = metricDescriptors{
		// the map key will function as an identifier of the metric throughout the rest of the code;
		// it is arbitrary, but by convention we use the actual metric name
		"resources":        NewMetricDesc("drbd", "resources", "The DRBD resources; 1 line per name, per volume", []string{"resource", "role", "volume", "disk_state"}),
		"connections":      NewMetricDesc("drbd", "connections", "The DRBD resource connections; 1 line per per resource, per peer_node_id", []string{"resource", "peer_node_id", "peer_role", "volume", "peer_disk_state"}),
		"connections_sync": NewMetricDesc("drbd", "connections_sync", "The in sync percentage value for DRBD resource connections", []string{"resource", "peer_node_id", "volume"}),
		"split_brain":      NewMetricDesc("drbd", "split_brain", "The split brain metric per node, resource and volume. The value 1 indicate the split brain is present", []string{"node", "resource", "volume"}),
	}
)

func NewDrbdCollector(drbdSetupPath string) (*drbdCollector, error) {
	err := CheckExecutables(drbdSetupPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize DRBD collector")
	}

	return &drbdCollector{
		DefaultCollector{
			metrics: drbdMetrics,
		},
		drbdSetupPath,
	}, nil
}

type drbdCollector struct {
	DefaultCollector
	drbdsetupPath string
}

func (c *drbdCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Infoln("Collecting DRBD metrics...")

	drbdStatusRaw, err := exec.Command(c.drbdsetupPath, "status", "--json").Output()
	if err != nil {
		log.Warnf("Error while retrieving drbd infos %s", err)
		return
	}
	// populate structs and parse relevant info we will expose via metrics
	drbdDev, err := parseDrbdStatus(drbdStatusRaw)
	if err != nil {
		log.Warnf("Error while parsing drbd json: %s", err)
		return
	}

	for _, resource := range drbdDev {
		for _, device := range resource.Devices {
			// the `resources` metric value is always 1, otherwise it's absent
			ch <- c.makeGaugeMetric("resources", float64(1), resource.Name, resource.Role, strconv.Itoa(device.Volume), strings.ToLower(device.DiskState))
		}
		if len(resource.Connections) == 0 {
			log.Warnf("Could not retrieve connection info for resource '%s'\n", resource.Name)
			continue
		}
		// a Resource can have multiple connection with different nodes
		for _, conn := range resource.Connections {
			if len(conn.PeerDevices) == 0 {
				log.Warnf("Could not retrieve any peer device info for connection '%d'\n", conn.PeerNodeID)
				continue
			}
			for _, peerDev := range conn.PeerDevices {
				ch <- c.makeGaugeMetric("connections", float64(1), resource.Name, strconv.Itoa(conn.PeerNodeID),
					conn.PeerRole, strconv.Itoa(peerDev.Volume), strings.ToLower(peerDev.PeerDiskState))

				ch <- c.makeGaugeMetric("connections_sync", float64(peerDev.PercentInSync), resource.Name, strconv.Itoa(conn.PeerNodeID), strconv.Itoa(peerDev.Volume))

			}
		}
	}

	// set split brain metric

	// 1) loop over location /var/lib/drbd/drbd-split-brain-detected-$DRBD_RESOURCE-$DRBD_VOLUME
	// 2) after prefix loop over-filenames and create metrics based on file name
	// 3) go throught the founded split-brain and set the metric accordingly (if there is no splitbrain just don't set to 0, as the overall design of exporter)

	// for foundedSplitBrains do
	// ch <- c.makeGaugeMetric("split_brain", float64(1), splitBrain["node"], splitBrain["resource"], splitBrain["volume"])

}

func parseDrbdStatus(statusRaw []byte) ([]drbdStatus, error) {
	var drbdDevs []drbdStatus
	err := json.Unmarshal(statusRaw, &drbdDevs)
	if err != nil {
		return drbdDevs, err
	}
	return drbdDevs, nil
}

// this function basically depends on the custom hook for drbd.
// by default if the custom hook is not set, the exporter will not be able to detect it.
// return a map of splitbrains
func parseDrbdSplitBrains() {
	log.Info("not yet")
}
