package main

import (
	"encoding/xml"
	"flag"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

type clusterMetrics struct {
	Node struct {
		Configured    int
		Online        int
		Standby       int
		StandbyOnFail int
		Maintenance   int
		Pending       int
		Unclean       int
		Shutdown      int
		ExpectedUp    int
		DC            int
		TypeMember    int
		TypePing      int
		TypeRemote    int
		TypeUnknown   int
	}
	Resource struct {
		Configured     int
		Unique         int
		Disabled       int
		Stopped        int
		Started        int
		Slave          int
		Master         int
		Active         int
		Orphaned       int
		Blocked        int
		Managed        int
		Failed         int
		FailureIgnored int
	}
}

// this historically from hawk-apiserver and parse some generic metrics
// it clusterStaterieve and parse cluster data and counters
func parseGenericMetrics(status *crmMon) *clusterMetrics {

	// clusterState save all the xml data . This is the metrics we will convert later to gauge etc.
	clusterState := &clusterMetrics{}
	clusterState.Node.Configured = status.Summary.Nodes.Number
	clusterState.Resource.Configured = status.Summary.Resources.Number
	clusterState.Resource.Disabled = status.Summary.Resources.Disabled
	rscIds := make(map[string]*resource)

	// Node informations
	for _, nod := range status.Nodes.Node {

		if nod.Online {
			clusterState.Node.Online++
		}
		if nod.Standby {
			clusterState.Node.Standby++
		}
		if nod.StandbyOnFail {
			clusterState.Node.StandbyOnFail++
		}
		if nod.Maintenance {
			clusterState.Node.Maintenance++
		}
		if nod.Pending {
			clusterState.Node.Pending++
		}
		if nod.Unclean {
			clusterState.Node.Unclean++
		}
		if nod.Shutdown {
			clusterState.Node.Shutdown++
		}
		if nod.ExpectedUp {
			clusterState.Node.ExpectedUp++
		}
		if nod.DC {
			clusterState.Node.DC++
		}
		if nod.Type == "member" {
			clusterState.Node.TypeMember++
		} else if nod.Type == "ping" {
			clusterState.Node.TypePing++
		} else if nod.Type == "remote" {
			clusterState.Node.TypeRemote++
		} else {
			clusterState.Node.TypeUnknown++
		}
		// node resources
		for _, rsc := range nod.Resources {
			rscIds[rsc.ID] = &rsc
			if rsc.Role == "Started" {
				clusterState.Resource.Started++
			} else if rsc.Role == "Stopped" {
				clusterState.Resource.Stopped++
			} else if rsc.Role == "Slave" {
				clusterState.Resource.Slave++
			} else if rsc.Role == "Master" {
				clusterState.Resource.Master++
			}
			if rsc.Active {
				clusterState.Resource.Active++
			}
			if rsc.Orphaned {
				clusterState.Resource.Orphaned++
			}
			if rsc.Blocked {
				clusterState.Resource.Blocked++
			}
			if rsc.Managed {
				clusterState.Resource.Managed++
			}
			if rsc.Failed {
				clusterState.Resource.Failed++
			}
			if rsc.FailureIgnored {
				clusterState.Resource.FailureIgnored++
			}
		}

	}
	clusterState.Resource.Unique = len(rscIds)
	return clusterState
}

var (
	// metrics with labels. (prefer these always as guideline)
	clusterNodes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_nodes",
			Help: "cluster nodes metrics for all of them",
		}, []string{"type"})

	nodeResources = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_node_resources",
			Help: "metric inherent per node resources",
		}, []string{"node", "resource_name", "role"})

	clusterResources = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_resources_status",
			Help: "global status of cluster resources",
		}, []string{"status"})
)

func initMetrics() {
	prometheus.MustRegister(clusterNodes)
	prometheus.MustRegister(nodeResources)
	prometheus.MustRegister(clusterResources)
}

var portNumber = flag.String("port", ":9002", "The port number to listen on for HTTP requests.")
var timeoutSeconds = flag.Int("timeout", 5, "timeout seconds for exporter to wait to fetch new data")

func main() {
	// read cli option and setup initial stat
	flag.Parse()
	initMetrics()
	http.Handle("/metrics", promhttp.Handler())

	// parse each X seconds the cluster configuration and update the metrics accordingly
	// this is done in a goroutine async. we update in this way each 2 second the metrics. (the second will be a parameter in future)
	go func() {
		for {
			// We want to reset certains metrics to 0 each time for removing the state.
			// since we have complex/nested metrics with multiples labels, unregistering/re-registering is the cleanest way.
			prometheus.Unregister(nodeResources)
			// overwrite metric with an empty one
			nodeResources = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "cluster_node_resources",
					Help: "metric inherent per node resources",
				}, []string{"node", "resource_name", "role"})
			err := prometheus.Register(nodeResources)
			if err != nil {
				log.Println("[ERROR]: failed to register NodeResource metric. Perhaps another exporter is already running?")
				log.Panic(err)
			}
			// get cluster status xml
			log.Println("[INFO]: Reading cluster configuration with crm_mon..")
			monxml, err := exec.Command("/usr/sbin/crm_mon", "-1", "--as-xml", "--group-by-node", "--inactive").Output()
			if err != nil {
				log.Println("[ERROR]: crm_mon command execution failed. Did you have crm_mon installed ?")
				log.Panic(err)
			}

			// read configuration
			var status crmMon
			err = xml.Unmarshal(monxml, &status)
			if err != nil {
				log.Println("[ERROR]: could not read cluster XML configuration")
				log.Panic(err)
			}

			metrics := parseGenericMetrics(&status)

			clusterResources.WithLabelValues("unique").Set(float64(metrics.Resource.Unique))
			clusterResources.WithLabelValues("disabled").Set(float64(metrics.Resource.Disabled))
			clusterResources.WithLabelValues("configured").Set(float64(metrics.Resource.Configured))
			clusterResources.WithLabelValues("active").Set(float64(metrics.Resource.Active))
			clusterResources.WithLabelValues("orpanhed").Set(float64(metrics.Resource.Orphaned))
			clusterResources.WithLabelValues("blocked").Set(float64(metrics.Resource.Blocked))
			clusterResources.WithLabelValues("managed").Set(float64(metrics.Resource.Managed))
			clusterResources.WithLabelValues("failed").Set(float64(metrics.Resource.Failed))
			clusterResources.WithLabelValues("failed_ignored").Set(float64(metrics.Resource.FailureIgnored))
			clusterResources.WithLabelValues("stopped").Set(float64(metrics.Resource.Stopped))
			clusterResources.WithLabelValues("started").Set(float64(metrics.Resource.Started))
			clusterResources.WithLabelValues("slave").Set(float64(metrics.Resource.Slave))
			clusterResources.WithLabelValues("master").Set(float64(metrics.Resource.Master))

			// nodes metrics
			clusterNodes.WithLabelValues("member").Set(float64(metrics.Node.TypeMember))
			clusterNodes.WithLabelValues("ping").Set(float64(metrics.Node.TypePing))
			clusterNodes.WithLabelValues("remote").Set(float64(metrics.Node.TypeRemote))
			clusterNodes.WithLabelValues("unknown").Set(float64(metrics.Node.TypeUnknown))
			clusterNodes.WithLabelValues("configured").Set(float64(metrics.Node.Configured))
			clusterNodes.WithLabelValues("online").Set(float64(metrics.Node.Online))
			clusterNodes.WithLabelValues("standby").Set(float64(metrics.Node.Standby))
			clusterNodes.WithLabelValues("standby_onfail").Set(float64(metrics.Node.StandbyOnFail))
			clusterNodes.WithLabelValues("maintenance").Set(float64(metrics.Node.Maintenance))
			clusterNodes.WithLabelValues("pending").Set(float64(metrics.Node.Pending))
			clusterNodes.WithLabelValues("unclean").Set(float64(metrics.Node.Unclean))
			clusterNodes.WithLabelValues("shutdown").Set(float64(metrics.Node.Shutdown))
			clusterNodes.WithLabelValues("expected_up").Set(float64(metrics.Node.ExpectedUp))
			clusterNodes.WithLabelValues("DC").Set(float64(metrics.Node.DC))

			// this produce a metric like: cluster_resources{node="dma-dog-hana01" resource_name="RA1"  role="master"} 1
			for _, nod := range status.Nodes.Node {
				for _, rsc := range nod.Resources {
					// increment if same resource is present
					nodeResources.WithLabelValues(strings.ToLower(nod.Name), strings.ToLower(rsc.ID), strings.ToLower(rsc.Role)).Inc()
				}
			}

			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
		}
	}()

	log.Println("[INFO]: Serving metrics on port", *portNumber)
	log.Println("[INFO]: refreshing metric timeouts set to", *timeoutSeconds)
	log.Fatal(http.ListenAndServe(*portNumber, nil))
}
