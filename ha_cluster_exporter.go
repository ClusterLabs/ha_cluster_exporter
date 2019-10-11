package main

import (
	"encoding/xml"
	"flag"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// this types are for reading pacemaker configuration xml when running crm_mon command
// and lookup the corrispective value
type crmMon struct {
	Version string  `xml:"version,attr"`
	Summary summary `xml:"summary"`
	Nodes   nodes   `xml:"nodes"`
}

type summary struct {
	Nodes struct {
		Number int `xml:"number,attr"`
	} `xml:"nodes_configured"`
	Resources resourcesConfigured `xml:"resources_configured"`
}

type resourcesConfigured struct {
	Number   int `xml:"number,attr"`
	Disabled int `xml:"disabled,attr"`
	Blocked  int `xml:"blocked,attr"`
}

type nodes struct {
	Node []node `xml:"node"`
}

type node struct {
	Name             string     `xml:"name,attr"`
	ID               string     `xml:"id,attr"`
	Online           bool       `xml:"online,attr"`
	Standby          bool       `xml:"standby,attr"`
	StandbyOnFail    bool       `xml:"standby_onfail,attr"`
	Maintenance      bool       `xml:"maintenance,attr"`
	Pending          bool       `xml:"pending,attr"`
	Unclean          bool       `xml:"unclean,attr"`
	Shutdown         bool       `xml:"shutdown,attr"`
	ExpectedUp       bool       `xml:"expected_up,attr"`
	DC               bool       `xml:"is_dc,attr"`
	ResourcesRunning int        `xml:"resources_running,attr"`
	Type             string     `xml:"type,attr"`
	Resources        []resource `xml:"resource"`
}

type resource struct {
	ID             string `xml:"id,attr"`
	Agent          string `xml:"resource_agent,attr"`
	Role           string `xml:"role,attr"`
	Active         bool   `xml:"active,attr"`
	Orphaned       bool   `xml:"orphaned,attr"`
	Blocked        bool   `xml:"blocked,attr"`
	Managed        bool   `xml:"managed,attr"`
	Failed         bool   `xml:"failed,attr"`
	FailureIgnored bool   `xml:"failure_ignored,attr"`
	NodesRunningOn int    `xml:"nodes_running_on,attr"`
}

var (
	// corosync metrics
	corosyncRingErrorsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "corosync_ring_errors_total",
		Help: "Total number of ring errors in corosync",
	})

	corosyncQuorate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "corosync_quorate",
		Help: "shows if the cluster is quorate. 1 cluster is quorate, 0 not",
	})

	// cluster metrics
	clusterNodesConf = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_configured_total",
		Help: "Number of nodes configured in ha cluster",
	})

	clusterResourcesConf = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_configured_total",
		Help: "Number of total configured resources in ha cluster",
	})

	// metrics with labels

	// sbd metrics
	sbdDevStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_sbd_device_status",
			Help: "cluster sbd status for each SBD device. 1 is healthy device, 0 is not",
		}, []string{"device_name"})

	// corosync quorum
	corosyncQuorum = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "corosync_quorum",
			Help: "cluster quorum information",
		}, []string{"type"})

	// cluster metrics
	clusterNodes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_nodes",
			Help: "cluster nodes metrics for all of them",
		}, []string{"node", "type"})

	nodeResources = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_node_resources",
			Help: "metric inherent per node resources",
		}, []string{"node", "resource_name", "role", "managed", "status"})

	// drbd metrics
	drbdDiskState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ha_cluster_drbd_resource_disk_state",
			Help: "disk states 0=Out of Sync, 1=UptoDate, 2=Syncing",
		}, []string{"resource_name", "role", "volume"})
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
			Name: "cluster_node_resources",
			Help: "metric inherent per node resources",
		}, []string{"node", "resource_name", "role", "managed", "status"})
	err := prometheus.Register(nodeResources)
	if err != nil {
		log.Println("[ERROR]: failed to register NodeResource metric. Perhaps another exporter is already running?")
		return err
	}

	prometheus.Unregister(clusterNodes)
	// overwrite metric with an empty one
	clusterNodes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_nodes",
			Help: "cluster nodes metrics for all of them",
		}, []string{"node", "type"})

	err = prometheus.Register(clusterNodes)
	if err != nil {
		log.Println("[ERROR]: failed to register clusterNode metric. Perhaps another exporter is already running?")
		return err
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
		for {
			// retrieve info via binary
			drbdStatusJSONRaw := getDrbdInfo()
			// populate structs and parse relevant info
			drbdDev, err := parseDrbdStatus(drbdStatusJSONRaw)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
				continue
			}
			// TODO DELETE metrics ! since we can't know if volume get destroyed

			// iterate over resources of drbd status
			for _, resource := range drbdDev {

				for _, device := range resource.Devices {
					// 0 out of sync, 1 UpToDate, 2 Syncing
					// lowercase the diskstates
					if "UpToDate" == resource.Devices[device.Volume].DiskState {
						drbdDiskState.WithLabelValues(resource.Name, resource.Role, strconv.Itoa(device.Volume)).Set(float64(1))
					}
					// TODO: check this better and lowercase it
					if "outofsync" == resource.Devices[device.Volume].DiskState {
						drbdDiskState.WithLabelValues(resource.Name, resource.Role, strconv.Itoa(device.Volume)).Set(float64(1))
					}
				}
			}
			// set drbd metric
			log.Println(drbdDev)
			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
		}
	}()
	// set SBD device metrics
	go func() {
		for {
			log.Println("[INFO]: Reading cluster SBD configuration..")
			// read configuration of SBD
			sbdConfiguration, err := readSdbFile()
			if err != nil {
				log.Println(err)
				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
				continue
			}
			// retrieve a list of sbd devices
			sbdDevices, err2 := getSbdDevices(sbdConfiguration)
			// mostly, the sbd_device were not set in conf file for returning an error
			if err2 != nil {
				log.Println(err2)
				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
				continue
			}
			// set and return a map of sbd devices with true healthy, false not
			sbdStatus := setSbdDeviceHealth(sbdDevices)

			if len(sbdStatus) == 0 {
				log.Println("[WARN]: Could not retrieve any sbd device")
				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
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

			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
		}
	}()

	// set corosync metrics: Ring errors
	go func() {
		for {
			ringStatus := getCorosyncRingStatus()
			ringErrorsTotal, err := parseRingStatus(ringStatus)
			if err != nil {
				log.Println("[ERROR]: could not execute command: usr/sbin/corosync-cfgtool -s")
				log.Println(err)
				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
				continue
			}
			corosyncRingErrorsTotal.Set(float64(ringErrorsTotal))
			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
		}
	}()
	// set corosync metrics: quorum metrics
	go func() {
		for {
			quoromStatus := getQuoromClusterInfo()
			voteQuorumInfo, quorate := parseQuoromStatus(quoromStatus)

			if len(voteQuorumInfo) == 0 {
				log.Println("[ERROR]: Could not retrieve any quorum information, map is empty")
				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
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
		for {

			// remove all global state contained by metrics
			err := resetClusterMetrics()
			if err != nil {
				log.Println("[ERROR]: fail to 	 reset metrics for cluster crm_mon component")
				log.Println(err)
				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
				continue
			}
			// get cluster status xml
			log.Println("[INFO]: Reading cluster configuration with crm_mon..")
			monxml, err := exec.Command("/usr/sbin/crm_mon", "-1", "--as-xml", "--group-by-node", "--inactive").Output()
			if err != nil {
				log.Println("[ERROR]: crm_mon command execution failed. Did you have crm_mon installed ?")
				log.Println(err)
				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
				continue
			}

			// read configuration
			var status crmMon
			err = xml.Unmarshal(monxml, &status)
			if err != nil {
				log.Println("[ERROR]: could not read cluster XML configuration")
				log.Println(err)
				time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
				continue
			}

			clusterResourcesConf.Set(float64(status.Summary.Resources.Number))
			clusterNodesConf.Set(float64(status.Summary.Nodes.Number))

			// set node metrics
			// cluster_nodes{node="dma-dog-hana01" type="master"} 1
			for _, nod := range status.Nodes.Node {
				if nod.Online {
					clusterNodes.WithLabelValues(nod.Name, "online").Set(float64(1))
				}
				if nod.Standby {
					clusterNodes.WithLabelValues(nod.Name, "standby").Set(float64(1))
				}
				if nod.StandbyOnFail {
					clusterNodes.WithLabelValues(nod.Name, "standby_onfail").Set(float64(1))
				}
				if nod.Maintenance {
					clusterNodes.WithLabelValues(nod.Name, "maintenance").Set(float64(1))
				}
				if nod.Pending {
					clusterNodes.WithLabelValues(nod.Name, "pending").Set(float64(1))
				}
				if nod.Unclean {
					clusterNodes.WithLabelValues(nod.Name, "unclean").Set(float64(1))
				}
				if nod.Shutdown {
					clusterNodes.WithLabelValues(nod.Name, "shutdown").Set(float64(1))
				}
				if nod.ExpectedUp {
					clusterNodes.WithLabelValues(nod.Name, "expected_up").Set(float64(1))
				}
				if nod.DC {
					clusterNodes.WithLabelValues(nod.Name, "dc").Set(float64(1))
				}
				if nod.Type == "member" {
					clusterNodes.WithLabelValues(nod.Name, "member").Set(float64(1))
				} else if nod.Type == "ping" {
					clusterNodes.WithLabelValues(nod.Name, "ping").Set(float64(1))
				} else if nod.Type == "remote" {
					clusterNodes.WithLabelValues(nod.Name, "remote").Set(float64(1))
				} else {
					clusterNodes.WithLabelValues(nod.Name, "unknown").Set(float64(1))
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

			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
		}
	}()

	log.Println("[INFO]: Serving metrics on port", *portNumber)
	log.Println("[INFO]: refreshing metric timeouts set to", *timeoutSeconds)
	log.Fatal(http.ListenAndServe(*portNumber, nil))
}
