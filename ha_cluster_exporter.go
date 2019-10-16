package main

import (
	"flag"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	// corosync metrics
	corosyncRingErrorsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ha_cluster_corosync_ring_errors_total",
		Help: "Total number of ring errors in corosync",
	})

	corosyncQuorate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ha_cluster_corosync_quorate",
		Help: "shows if the cluster is quorate. 1 cluster is quorate, 0 not",
	})

	// cluster metrics
	clusterNodesConf = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ha_cluster_nodes_configured_total",
		Help: "Number of nodes configured in ha cluster",
	})

	clusterResourcesConf = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ha_cluster_resources_configured_total",
		Help: "Number of total configured resources in ha cluster",
	})

	// metrics with labels

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

	// cluster metrics
	clusterNodes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_nodes",
			Help: "cluster nodes metrics for all of them",
		}, []string{"node_name", "type"})

	nodeResources = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_node_resources",
			Help: "metric inherent per node resources",
		}, []string{"node_name", "resource_name", "role", "managed", "status"})

	// drbd metrics
	drbdDiskState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_drbd_resource",
			Help: "show per resource name, its role, the volume and disk_state (Diskless,Attaching, Failed, Negotiating, Inconsistent, Outdated, DUnknown, Consistent, UpToDate)",
		}, []string{"resource_name", "role", "volume", "disk_state"})

	drbdRemoteDiskState = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ha_cluster_drbd_resource_remote_connection",
				Help: "show per remote connection resource name, its role, the volume and disk_state (Diskless,Attaching, Failed, Negotiating, Inconsistent, Outdated, DUnknown, Consistent, UpToDate)",
			}, []string{"resource_name", "peer-role", "volume", "peer-node-id", "peer-disk-state" })
)

func init() {
	prometheus.MustRegister(clusterNodes)
	prometheus.MustRegister(nodeResources)
	prometheus.MustRegister(clusterResourcesConf)
	prometheus.MustRegister(clusterNodesConf)
	prometheus.MustRegister(corosyncRingErrorsTotal)
	prometheus.MustRegister(corosyncQuorum)
	prometheus.MustRegister(corosyncQuorate)
	prometheus.MustRegister(sbdDevStatus)
	prometheus.MustRegister(drbdDiskState)
	prometheus.MustRegister(drbdRemoteDiskState)
}

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
	prometheus.Unregister(drbdRemoteDiskState)
	drbdRemoteDiskState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_drbd_resource_remote_connection",
			Help: "show per remote connection resource name, its role, the volume and disk_state (Diskless,Attaching, Failed, Negotiating, Inconsistent, Outdated, DUnknown, Consistent, UpToDate)",
		}, []string{"resource_name", "peer-role", "volume", "peer-node-id", "peer-disk-state" })
	if err != nil {
			return errors.Wrap(err, "failed to register DRBD remote disk state metric. Perhaps another exporter is already running?")
	}
	return nil
}

var portNumber = flag.String("port", ":9002", "The port number to listen on for HTTP requests.")
var timeoutSeconds = flag.Int("timeout", 5, "timeout seconds for exporter to wait to fetch new data")

func main() {
	// read cli option and setup initial stat
	flag.Parse()
	http.Handle("/metrics", promhttp.Handler())

	// for each different metrics, handle it in differents gorutines, and use same timeout.

	// set DRBD metrics

	go func() {
		log.Infoln("Starting DRBD metrics collector...")


		for {
			
			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
		
			

			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)

			log.Infoln("Reading DRBD status...")

			// retrieve drbdInfos calling its binary
			drbdStatusJSONRaw, err := getDrbdInfo()
			if err != nil {
				log.Errorln(err)
				continue
			}
			// populate structs and parse relevant info we will expose via metrics
			drbdDev, err := parseDrbdStatus(drbdStatusJSONRaw)
			if err != nil {
				log.Errorln(err)
				continue
			}

			// reset metrics before setting news to remove any state information
			err = resetDrbdMetrics()
			if err != nil {
				log.Errorln(err)
				continue
			}

			// create a metric like : ha_cluster_drbd_resource{resource_name="1-single-0", role="primary", volume="0",  disk_state="uptodate"} 1
			// the metric is always set to 1 or is absent
			for _, resource := range drbdDev {
				for _, device := range resource.Devices {
					drbdDiskState.WithLabelValues(resource.Name, resource.Role, strconv.Itoa(device.Volume), strings.ToLower(resource.Devices[device.Volume].DiskState)).Set(float64(1))
				}
			}
		}
	}()

	// set SBD device metrics
	go func() {
		if _, err := os.Stat("/etc/sysconfig/sbd"); os.IsNotExist(err) {
			log.Warnln("SBD configuration not available, SBD metrics won't be collected")
			return
		}

		log.Infoln("Starting SBD metrics collector...")


		for {
		
			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
	
		
			// read configuration of SBD
			sbdConfiguration, err := readSdbFile()
			if err != nil {
				log.Errorln(err)
				continue
			}
			// retrieve a list of sbd devices
			sbdDevices, err := getSbdDevices(sbdConfiguration)
			// mostly, the sbd_device were not set in conf file for returning an error
			if err != nil {
				log.Errorln(err)
				continue
			}

			// set and return a map of sbd devices with true if healthy, false if not
			sbdStatus, err := getSbdDeviceHealth(sbdDevices)
			if err != nil {
				log.Errorln(err)
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

		for {
			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
					
			log.Infoln("Reading ring status...")
			ringStatus := getCorosyncRingStatus()
			ringErrorsTotal, err := parseRingStatus(ringStatus)
			if err != nil {
				log.Errorln(err)
				continue
			}
			corosyncRingErrorsTotal.Set(float64(ringErrorsTotal))
		}
	}()

	// set corosync metrics: quorum metrics
	go func() {
		log.Infoln("Starting corosync quorum metrics collector...")

		for {
			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
		
			

			log.Infoln("Reading quorum status...")
			quoromStatus := getQuoromClusterInfo()
			voteQuorumInfo, quorate, err := parseQuoromStatus(quoromStatus)
			if err != nil {
				log.Errorln(err)
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

	// set cluster pacemaker metrics
	go func() {
		log.Infoln("Starting pacemaker metrics collector...")

	
		for {
		
			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
		

			// remove all global state contained by metrics
			err := resetClusterMetrics()
			if err != nil {
				log.Errorln(err)
				continue
			}

			// get cluster status xml
			log.Infoln("Reading cluster configuration with crm_mon..")
			pacemakerXMLRaw, err := exec.Command("/usr/sbin/crm_mon", "-1", "--as-xml", "--group-by-node", "--inactive").Output()
			if err != nil {
				log.Errorln(err)
				continue
			}

			// parse raw XML returned from crm_mon and populate structs for metrics
			status, err := parsePacemakerStatus(pacemakerXMLRaw)
			if err != nil {
				log.Errorln(err)
				continue
			}

			clusterResourcesConf.Set(float64(status.Summary.Resources.Number))
			clusterNodesConf.Set(float64(status.Summary.Nodes.Number))

			// set node metrics
			// cluster_nodes{node="dma-dog-hana01" type="master"} 1
			for _, node := range status.Nodes.Node {
				if node.Online {
					clusterNodes.WithLabelValues(node.Name, "online").Set(float64(1))
				}
				if node.Standby {
					clusterNodes.WithLabelValues(node.Name, "standby").Set(float64(1))
				}
				if node.StandbyOnFail {
					clusterNodes.WithLabelValues(node.Name, "standby_onfail").Set(float64(1))
				}
				if node.Maintenance {
					clusterNodes.WithLabelValues(node.Name, "maintenance").Set(float64(1))
				}
				if node.Pending {
					clusterNodes.WithLabelValues(node.Name, "pending").Set(float64(1))
				}
				if node.Unclean {
					clusterNodes.WithLabelValues(node.Name, "unclean").Set(float64(1))
				}
				if node.Shutdown {
					clusterNodes.WithLabelValues(node.Name, "shutdown").Set(float64(1))
				}
				if node.ExpectedUp {
					clusterNodes.WithLabelValues(node.Name, "expected_up").Set(float64(1))
				}
				if node.DC {
					clusterNodes.WithLabelValues(node.Name, "dc").Set(float64(1))
				}
				if node.Type == "member" {
					clusterNodes.WithLabelValues(node.Name, "member").Set(float64(1))
				} else if node.Type == "ping" {
					clusterNodes.WithLabelValues(node.Name, "ping").Set(float64(1))
				} else if node.Type == "remote" {
					clusterNodes.WithLabelValues(node.Name, "remote").Set(float64(1))
				} else {
					clusterNodes.WithLabelValues(node.Name, "unknown").Set(float64(1))
				}
			}

			// parse node status
			// this produce a metric like:
			//	cluster_node_resources{managed="false",node="dma-dog-hana01",resource_name="rsc_saphanatopology_prd_hdb00",role="started",status="active"} 1
			//  cluster_node_resources{managed="true",node="dma-dog-hana01",resource_name="rsc_ip_prd_hdb00",role="started",status="active"} 1
			for _, nod := range status.Nodes.Node {
				for _, rsc := range nod.Resources {
					if rsc.Active {
						nodeResources.WithLabelValues(strings.ToLower(nod.Name), strings.ToLower(rsc.ID), strings.ToLower(rsc.Role), strconv.FormatBool(rsc.Managed),
							"active").Inc()
					}
					if rsc.Orphaned {
						nodeResources.WithLabelValues(strings.ToLower(nod.Name), strings.ToLower(rsc.ID), strings.ToLower(rsc.Role), strconv.FormatBool(rsc.Managed),
							"orphaned").Inc()
					}
					if rsc.Blocked {
						nodeResources.WithLabelValues(strings.ToLower(nod.Name), strings.ToLower(rsc.ID), strings.ToLower(rsc.Role), strconv.FormatBool(rsc.Managed),
							"blocked").Inc()
					}
					if rsc.Failed {
						nodeResources.WithLabelValues(strings.ToLower(nod.Name), strings.ToLower(rsc.ID), strings.ToLower(rsc.Role), strconv.FormatBool(rsc.Managed),
							"failed").Inc()
					}
					if rsc.FailureIgnored {
						nodeResources.WithLabelValues(strings.ToLower(nod.Name), strings.ToLower(rsc.ID), strings.ToLower(rsc.Role), strconv.FormatBool(rsc.Managed),
							"failed_ignored").Inc()
					}
				}

			}
		}
	}()

	log.Infoln("Serving metrics on port", *portNumber)
	log.Infoln("refreshing metric timeouts set to", *timeoutSeconds)
	log.Fatal(http.ListenAndServe(*portNumber, nil))
}
