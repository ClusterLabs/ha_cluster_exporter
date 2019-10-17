package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const NAMESPACE = "ha_cluster"

type metricsGroup map[string]*prometheus.Desc

var (
/*
	// corosync metrics
	corosyncRingErrorsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ha_cluster_corosync_ring_errors_total",
		Help: "Total number of ring errors in corosync",
	})

	corosyncQuorate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ha_cluster_corosync_quorate",
		Help: "shows if the cluster is quorate. 1 cluster is quorate, 0 not",
	})
	// sbd metrics
	sbdDevStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_sbd_device_status",
			Help: "cluster sbd status for each SBD device. 1 is healthy device, 0 is not",
		}, []string{"device_name"})

	// corosync quorum
	corosyncQuorum = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_corosync_quorum",
			Help: "cluster quorum information",
		}, []string{"type"})

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

func newMetricDesc(subsystem, name, help string, variableLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(NAMESPACE, subsystem, name), help, variableLabels, nil)
}

func init() {
	/*
		prometheus.MustRegister(corosyncRingErrorsTotal)
		prometheus.MustRegister(corosyncQuorum)
		prometheus.MustRegister(corosyncQuorate)
		prometheus.MustRegister(sbdDevStatus)
		prometheus.MustRegister(drbdDiskState)
		prometheus.MustRegister(remoteDrbdDiskState)*/
	pacemakerCollector, err := NewPacemakerCollector("/usr/sbin/crm_mon")
	if err != nil {
		log.Warnln(err)
	} else {
		prometheus.MustRegister(pacemakerCollector)
	}
}

/*
// this function is for some cluster metrics which have resource as labels.
// since we cannot be sure a resource exists always, we need to destroy the metrics at each iteration
// otherwise we will have wrong metrics ( thinking a resource exist when not)
func resetClusterMetrics() error {
	// We want to reset certains metrics to 0 each time for removing the state.
	// since we have complex/nested metrics with multiples labels, unregistering/re-registering is the cleanest way.
	prometheus.Unregister(nodeResources)
	// overwrite metric with an empty one
	nodeResources = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_node_resources",
			Help: "metric inherent per node resources",
		}, []string{"node_name", "resource_name", "role", "managed", "status"})
	err := prometheus.Register(nodeResources)
	if err != nil {
		return errors.Wrap(err, "failed to register NodeResource metric. Perhaps another exporter is already running?")
	}

	prometheus.Unregister(clusterNodes)
	// overwrite metric with an empty one
	clusterNodes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_nodes",
			Help: "cluster nodes metrics for all of them",
		}, []string{"node_name", "type"})

	err = prometheus.Register(clusterNodes)
	if err != nil {
		return errors.Wrap(err, "failed to register clusterNode metric. Perhaps another exporter is already running?")
	}
	return nil
}

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

		// set SBD device metrics
		go func() {
			firstTime := true
			if _, err := os.Stat("/etc/sysconfig/sbd"); os.IsNotExist(err) {
				log.Warnln("SBD configuration not available, SBD metrics won't be collected")
				return
			}

			log.Infoln("Starting SBD metrics collector...")

			for {
				sleepDefaultTimeout(&firstTime)
				// read configuration of SBD
				sbdConfiguration, err := readSdbFile()
				if err != nil {
					log.Warnln(err)
					continue
				}
				// retrieve a list of sbd devices
				sbdDevices, err := getSbdDevices(sbdConfiguration)
				// mostly, the sbd_device were not set in conf file for returning an error
				if err != nil {
					log.Warnln(err)
					continue
				}

				// set and return a map of sbd devices with true if healthy, false if not
				sbdStatus, err := getSbdDeviceHealth(sbdDevices)
				if err != nil {
					log.Warnln(err)
					continue
				}
				for sbdDev, sbdStatusBool := range sbdStatus {
					// true it means the sbd device is healthy
					if sbdStatusBool == true {
						sbdDevStatus.WithLabelValues(sbdDev).Set(float64(1))
					} else {
						sbdDevStatus.WithLabelValues(sbdDev).Set(float64(0))
					}
				}
			}
		}()

		// set corosync metrics: ring errors
		go func() {
			log.Infoln("Starting corosync ring errors collector...")
			firstTime := true
			for {

				sleepDefaultTimeout(&firstTime)

				log.Infoln("Reading ring status...")
				ringStatus := getCorosyncRingStatus()
				ringErrorsTotal, err := parseRingStatus(ringStatus)
				if err != nil {
					log.Warnln(err)
					continue
				}
				corosyncRingErrorsTotal.Set(float64(ringErrorsTotal))
			}
		}()

		// set corosync metrics: quorum metrics
		go func() {
			log.Infoln("Starting corosync quorum metrics collector...")
			firstTime := true
			for {
				sleepDefaultTimeout(&firstTime)

				log.Infoln("Reading quorum status...")
				quoromStatus, err := getQuoromClusterInfo()
				if err != nil {
					log.Warnln(err)
					continue
				}
				voteQuorumInfo, quorate, err := parseQuoromStatus(quoromStatus)
				if err != nil {
					log.Warnln(err)
					continue
				}

				// set metrics relative to quorum infos
				corosyncQuorum.WithLabelValues("expected_votes").Set(float64(voteQuorumInfo["expectedVotes"]))
				corosyncQuorum.WithLabelValues("highest_expected").Set(float64(voteQuorumInfo["highestExpected"]))
				corosyncQuorum.WithLabelValues("total_votes").Set(float64(voteQuorumInfo["totalVotes"]))
				corosyncQuorum.WithLabelValues("quorum").Set(float64(voteQuorumInfo["quorum"]))

				// set metric if we have a quorate or not
				// 1 means we have it
				if quorate == "yes" {
					corosyncQuorate.Set(float64(1))
				}

				if quorate == "no" {
					corosyncQuorate.Set(float64(0))
				}

				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
			}
		}()

	*/
	log.Infoln("Serving metrics on port", *portNumber)
	log.Fatal(http.ListenAndServe(*portNumber, nil))
}
