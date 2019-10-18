package main

import (
	"encoding/json"
	"os"
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
			Volume        int    `json:"volume"`
			PeerDiskState string `json:"peer-disk-state"`
		} `json:"peer_devices"`
	} `json:"connections"`
}

var (
	drbdMetrics = metricDescriptors{
		// the map key will function as an identifier of the metric throughout the rest of the code;
		// it is arbitrary, but by convention we use the actual metric name
		"resources": NewMetricDesc(
			"drbd",
			"resources",
			"Describes each DRBD resource; 1 line per resource_name",
			[]string{"resource_name", "role", "volume", "disk_state"}),
		"resources_remote_connections": NewMetricDesc(
			"drbd",
			"resources_remote_connections",
			"Describes each DRBD resource connection; 1 line per peer device",
			[]string{"resource_name", "peer_node_id", "peer_role", "volume", "peer_disk_state"}),
	}
	drbdsetupPath = "/usr/sbin/drbdsetup"
)

func NewDrbdCollector() (*drbdCollector, error) {
	fileInfo, err := os.Stat(drbdsetupPath)
	if err != nil || os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "'%s' not found", drbdsetupPath)
	}
	if (fileInfo.Mode() & 0111) == 0 {
		return nil, errors.Errorf("'%s' is not executable", drbdsetupPath)
	}

	return &drbdCollector{
		DefaultCollector{
			metrics: drbdMetrics,
		},
	}, nil
}

type drbdCollector struct {
	DefaultCollector
}

func (c *drbdCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Infoln("Collecting DRBD metrics...")

	drbdStatusJSONRaw, err := getDrbdInfo()
	if err != nil {
		log.Warnf("Error by retrieving drbd infos %s", err)
		return
	}
	// populate structs and parse relevant info we will expose via metrics
	drbdDev, err := parseDrbdStatus(drbdStatusJSONRaw)
	if err != nil {
		log.Warnf("Error by parsing drbd json: %s", err)
		return
	}

	// the metric is always set to 1 or is absent
	for _, resource := range drbdDev {
		for _, device := range resource.Devices {
			ch <- c.makeGaugeMetric("resources", float64(1), resource.Name, resource.Role, strconv.Itoa(device.Volume), strings.ToLower(resource.Devices[device.Volume].DiskState))
		}
		// 2) ha_cluster_drbd_resource_remote_connection{resource_name="1-single-0", peer_node_id="1", role="primary", volume="0",  disk_state="uptodate"} 1
		// a resource could not have any connection
		if len(resource.Connections) == 0 {
			log.Warnf("Could not retrieve connection info for resource '%s'\n", resource.Name)
			continue
		}
		// a Resource can have multiple connection with different nodes
		for _, conn := range resource.Connections {
			// pro resource go through the volume of peer and its peer state
			if len(conn.PeerDevices) == 0 {
				log.Warnf("Could not retrieve any peer device info for connection '%d'\n", conn.PeerNodeID)
				continue
			}
			for _, peerDev := range conn.PeerDevices {
				ch <- c.makeGaugeMetric("resources_remote_connections", float64(1), resource.Name, strconv.Itoa(conn.PeerNodeID), conn.PeerRole, strconv.Itoa(peerDev.Volume), strings.ToLower(peerDev.PeerDiskState))
			}
		}
	}
}

// return drbd status in byte raw json
func getDrbdInfo() ([]byte, error) {
	drbdStatusRaw, err := exec.Command(drbdsetupPath, "status", "--json").Output()
	return drbdStatusRaw, err
}

func parseDrbdStatus(statusRaw []byte) ([]drbdStatus, error) {
	var drbdDevs []drbdStatus
	err := json.Unmarshal(statusRaw, &drbdDevs)
	if err != nil {
		return drbdDevs, err
	}
	return drbdDevs, nil
}
