package main

import (
	"flag"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const NAMESPACE = "ha_cluster"

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

type metricDescriptors map[string]*prometheus.Desc

type DefaultCollector struct {
	metrics metricDescriptors
	mutex   sync.RWMutex
}

func (c *DefaultCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *DefaultCollector) makeMetric(metricKey string, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	desc, ok := c.metrics[metricKey]
	if !ok {
		// we hard panic on this because it's most certainly a coding error
		panic(errors.Errorf("undeclared metric '%s'", metricKey))
	}
	return prometheus.NewMetricWithTimestamp(clock.Now(), prometheus.MustNewConstMetric(desc, valueType, value, labelValues...))
}

func NewMetricDesc(subsystem, name, help string, variableLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(NAMESPACE, subsystem, name), help, variableLabels, nil)
}

var (
	clock Clock = &SystemClock{}

/*
	// drbd metrics
	drbdDiskState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_drbd_resource",
			Help: "show per resource name, its role, the volume and disk_state (Diskless,Attaching, Failed, Negotiating, Inconsistent, Outdated, DUnknown, Consistent, UpToDate)",
		}, []string{"resource_name", "role", "volume", "disk_state"})

	remoteDrbdDiskState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_drbd_resources_remote_connection",
			Help: "show per remote connection resource name, its role, peer-id , the volume and disk_state (Diskless,Attaching, Failed, Negotiating, Inconsistent, Outdated, DUnknown, Consistent, UpToDate)",
		}, []string{"resource_name", "peer_node_id", "peer_role", "volume", "peer_disk_state"})
*/
)

/*

func resetDrbdMetrics() error {
	// for Drbd we need to reset remove metrics state because some disk could be removed during cluster lifecycle
	// so we need to have a clean atomic snapshot
	prometheus.Unregister(drbdDiskState)
	// overwrite metric with an empty one
	drbdDiskState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_drbd_resources",
			Help: "show per resource name, its role, the volume and disk_state (Diskless,Attaching, Failed, Negotiating, Inconsistent, Outdated, DUnknown, Consistent, UpToDate)",
		}, []string{"resource_name", "role", "volume", "disk_state"})
	err := prometheus.Register(drbdDiskState)
	if err != nil {
		return errors.Wrap(err, "failed to register DRBD disk state metric. Perhaps another exporter is already running?")
	}
	prometheus.Unregister(remoteDrbdDiskState)
	remoteDrbdDiskState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_drbd_resources_remote_connection",
			Help: "show per remote connection resource name, its role, peer-id , the volume and disk_state (Diskless,Attaching, Failed, Negotiating, Inconsistent, Outdated, DUnknown, Consistent, UpToDate)",
		}, []string{"resource_name", "peer_node_id", "peer_role", "volume", "peer_disk_state"})
	err = prometheus.Register(remoteDrbdDiskState)
	if err != nil {
		return errors.Wrap(err, "failed to register DRBD remote disk state metric. Perhaps another exporter is already running?")
	}
	return nil
}

// helper function for setting timeout for goroutines
// we skip the timeout by init apply each time later
func sleepDefaultTimeout(firstTime *bool) {
	// by the initialization of exporter we don't wait timeout but serve metrics.
	if *firstTime {
		*firstTime = false
		return
	}
	time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
}
*/
var portNumber = flag.String("port", ":9002", "The port number to listen on for HTTP requests.")

func main() {
	// read cli option and setup initial stat
	flag.Parse()

	pacemakerCollector, err := NewPacemakerCollector()
	if err != nil {
		log.Warnln(err)
	} else {
		prometheus.MustRegister(pacemakerCollector)
	}

	corosyncCollector, err := NewCorosyncCollector()
	if err != nil {
		log.Warnln(err)
	} else {
		prometheus.MustRegister(corosyncCollector)
	}

	sbdCollector, err := NewSbdCollector()
	if err != nil {
		log.Warnln(err)
	} else {
		prometheus.MustRegister(sbdCollector)
	}

	http.Handle("/metrics", promhttp.Handler())

	// for each different metrics, handle it in differents gorutines, and use same timeout.

	/*
		// set DRBD metrics
		go func() {
			if _, err := os.Stat("/sbin/drbdsetup"); os.IsNotExist(err) {
				log.Warnln("drbdsetup binary not available, DRBD metrics won't be collected")
				return
			}

			log.Infoln("Starting DRBD metrics collector...")
			firstTime := true
			for {
				sleepDefaultTimeout(&firstTime)
				log.Infoln("Reading DRBD status...")

				// retrieve drbdInfos calling its binary
				drbdStatusJSONRaw, err := getDrbdInfo()
				if err != nil {
					log.Warnf("Error by retrieving drbd infos %s", err)
					continue
				}
				// populate structs and parse relevant info we will expose via metrics
				drbdDev, err := parseDrbdStatus(drbdStatusJSONRaw)
				if err != nil {
					log.Warnf("Error by parsing drbd json: %s", err)
					continue
				}

				// reset metrics before setting news to remove any state information
				err = resetDrbdMetrics()
				if err != nil {
					log.Warnf("Error by resetting drbd metrics %s", err)
					continue
				}

				// 1) ha_cluster_drbd_resource{resource_name="1-single-0", role="primary", volume="0",  disk_state="uptodate"} 1
				// the metric is always set to 1 or is absent
				for _, resource := range drbdDev {
					for _, device := range resource.Devices {
						drbdDiskState.WithLabelValues(resource.Name, resource.Role, strconv.Itoa(device.Volume), strings.ToLower(resource.Devices[device.Volume].DiskState)).Set(float64(1))
					}
					// 2) ha_cluster_drbd_resource_remote_connection{resource_name="1-single-0", peer_node_id="1", role="primary", volume="0",  disk_state="uptodate"} 1
					// a resource could not have any connection
					if len(resource.Connections) == 0 {
						log.Warnln("could not retrieve any remote disk state connection info")
						continue
					}
					// a Resource can have multiple connection with different nodes
					for _, conn := range resource.Connections {
						// []string{"resource_name", "peer_node_id", "peer_role", "volume", "peer_disk_state"})
						// pro resource go through the volume of peer and its peer state
						if len(conn.PeerDevices) == 0 {
							log.Warnln("could not retrieve any peer Devices metric")
							continue
						}
						for _, peerDev := range conn.PeerDevices {
							remoteDrbdDiskState.WithLabelValues(resource.Name, strconv.Itoa(conn.PeerNodeID), conn.PeerRole, strconv.Itoa(peerDev.Volume), strings.ToLower(peerDev.PeerDiskState)).Set(float64(1))
						}
					}
				}
			}
		}()

	*/
	log.Infoln("Serving metrics on port", *portNumber)
	log.Fatal(http.ListenAndServe(*portNumber, nil))
}
