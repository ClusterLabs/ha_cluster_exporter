package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sort"
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
	PerNode map[string]perNodeMetrics
}

type perNodeMetrics struct {
	ResourcesRunning int
}

// this historically from hawk-apiserver and parse some generic metrics
func parseGenericMetrics(status *crmMon) *clusterMetrics {
	ret := &clusterMetrics{}

	ret.Node.Configured = status.Summary.Nodes.Number
	ret.Resource.Configured = status.Summary.Resources.Number
	ret.Resource.Disabled = status.Summary.Resources.Disabled
	ret.PerNode = make(map[string]perNodeMetrics)

	rscIds := make(map[string]*resource)

	for _, nod := range status.Nodes.Node {
		perNode := perNodeMetrics{ResourcesRunning: nod.ResourcesRunning}
		ret.PerNode[nod.Name] = perNode

		if nod.Online {
			ret.Node.Online++
		}
		if nod.Standby {
			ret.Node.Standby++
		}
		if nod.StandbyOnFail {
			ret.Node.StandbyOnFail++
		}
		if nod.Maintenance {
			ret.Node.Maintenance++
		}
		if nod.Pending {
			ret.Node.Pending++
		}
		if nod.Unclean {
			ret.Node.Unclean++
		}
		if nod.Shutdown {
			ret.Node.Shutdown++
		}
		if nod.ExpectedUp {
			ret.Node.ExpectedUp++
		}
		if nod.DC {
			ret.Node.DC++
		}
		if nod.Type == "member" {
			ret.Node.TypeMember++
		} else if nod.Type == "ping" {
			ret.Node.TypePing++
		} else if nod.Type == "remote" {
			ret.Node.TypeRemote++
		} else {
			ret.Node.TypeUnknown++
		}

		for _, rsc := range nod.Resources {
			rscIds[rsc.ID] = &rsc
			if rsc.Role == "Started" {
				ret.Resource.Started++
			} else if rsc.Role == "Stopped" {
				ret.Resource.Stopped++
			} else if rsc.Role == "Slave" {
				ret.Resource.Slave++
			} else if rsc.Role == "Master" {
				ret.Resource.Master++
			}
			if rsc.Active {
				ret.Resource.Active++
			}
			if rsc.Orphaned {
				ret.Resource.Orphaned++
			}
			if rsc.Blocked {
				ret.Resource.Blocked++
			}
			if rsc.Managed {
				ret.Resource.Managed++
			}
			if rsc.Failed {
				ret.Resource.Failed++
			}
			if rsc.FailureIgnored {
				ret.Resource.FailureIgnored++
			}
		}
	}

	ret.Resource.Unique = len(rscIds)

	return ret
}

var (
	// simple gauge metric
	clusterNodesConf = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_configured",
		Help: "Number of nodes configured in ha cluster",
	})

	clusterNodesOnline = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_online",
		Help: "Number of nodes online in ha cluster",
	})

	clusterNodesStandby = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_standby",
		Help: "Number of nodes standby in ha cluster",
	})

	clusterNodesStandbyOnFail = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_stanby_onfail",
		Help: "Number of nodes standby onfail in ha cluster",
	})

	clusterNodesMaintenance = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_maintenance",
		Help: "Number of nodes in maintainance in ha cluster",
	})

	clusterNodesPending = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_pending",
		Help: "Number of nodes pending in ha cluster",
	})

	clusterNodesUnclean = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_unclean",
		Help: "Number of nodes unclean in ha cluster",
	})

	clusterNodesShutdown = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_shutdown",
		Help: "Number of nodes shutdown in ha cluster",
	})

	clusterNodesExpectedUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_expected_up",
		Help: "Number of nodes expected up in ha cluster",
	})

	clusterNodesDC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_expected_dc",
		Help: "Number of nodes dc in ha cluster",
	})

	// a gauge metric with label
	clusterNodes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_nodes",
			Help: "cluster nodes metrics",
		}, []string{"type"})

	clusterResourcesRunning = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_resources_running",
			Help: "number of cluster resources running",
		}, []string{"node"})

	clusterResources = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_resources",
			Help: "number of cluster resources",
		}, []string{"node", "resource_name", "role"})

	clusterResourcesStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_resources_status",
			Help: "status of cluster resources",
		}, []string{"status"})
)

func initMetrics() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(clusterNodesConf)
	prometheus.MustRegister(clusterNodesOnline)
	prometheus.MustRegister(clusterNodesStandby)
	prometheus.MustRegister(clusterNodesStandbyOnFail)
	prometheus.MustRegister(clusterNodesMaintenance)
	prometheus.MustRegister(clusterNodesPending)
	prometheus.MustRegister(clusterNodesUnclean)
	prometheus.MustRegister(clusterNodesShutdown)
	prometheus.MustRegister(clusterNodesExpectedUp)
	prometheus.MustRegister(clusterNodesDC)

	// metrics with labels
	prometheus.MustRegister(clusterNodes)
	prometheus.MustRegister(clusterResourcesRunning)
	prometheus.MustRegister(clusterResources)
	prometheus.MustRegister(clusterResourcesStatus)

}

var portNumber = flag.String("port", ":9001", "The port number to listen on for HTTP requests.")

func main() {
	// read cli option and setup initial stat
	flag.Parse()
	initMetrics()
	http.Handle("/metrics", promhttp.Handler())

	// parse each 2 seconds the cluster configuration and update the metrics accordingly
	// this is done in a goroutine async. we update in this way each 2 second the metrics. (the second will be a parameter in future)
	go func() {

		for {

			var status crmMon
			// get cluster status xml
			fmt.Println("[INFO]: Reading cluster configuration with crm_mon..")
			monxml, err := exec.Command("/usr/sbin/crm_mon", "-1", "--as-xml", "--group-by-node", "--inactive").Output()
			if err != nil {
				fmt.Println("[ERROR]: crm_mon command was not executed correctly. Did you have crm_mon installed ?")
				panic(err)
			}

			// read configuration
			err = xml.Unmarshal(monxml, &status)
			if err != nil {
				panic(err)
			}

			metrics := parseGenericMetrics(&status)
			// add genric node metrics
			clusterNodesConf.Set(float64(metrics.Node.Configured))
			clusterNodesOnline.Set(float64(metrics.Node.Online))
			clusterNodesStandby.Set(float64(metrics.Node.Standby))
			clusterNodesStandbyOnFail.Set(float64(metrics.Node.StandbyOnFail))
			clusterNodesMaintenance.Set(float64(metrics.Node.Maintenance))
			clusterNodesPending.Set(float64(metrics.Node.Pending))
			clusterNodesUnclean.Set(float64(metrics.Node.Unclean))
			clusterNodesShutdown.Set(float64(metrics.Node.Shutdown))
			clusterNodesExpectedUp.Set(float64(metrics.Node.ExpectedUp))
			clusterNodesDC.Set(float64(metrics.Node.DC))

			// ressouce status metrics
			clusterResourcesStatus.WithLabelValues("unique").Set(float64(metrics.Resource.Unique))
			clusterResourcesStatus.WithLabelValues("disabled").Set(float64(metrics.Resource.Disabled))
			clusterResourcesStatus.WithLabelValues("configured").Set(float64(metrics.Resource.Configured))
			clusterResourcesStatus.WithLabelValues("active").Set(float64(metrics.Resource.Active))
			clusterResourcesStatus.WithLabelValues("orpanhed").Set(float64(metrics.Resource.Orphaned))
			clusterResourcesStatus.WithLabelValues("blocked").Set(float64(metrics.Resource.Blocked))
			clusterResourcesStatus.WithLabelValues("managed").Set(float64(metrics.Resource.Managed))
			clusterResourcesStatus.WithLabelValues("failed").Set(float64(metrics.Resource.Failed))
			clusterResourcesStatus.WithLabelValues("failed_ignored").Set(float64(metrics.Resource.FailureIgnored))

			// metrics with labels
			clusterNodes.WithLabelValues("member").Set(float64(metrics.Node.TypeMember))
			clusterNodes.WithLabelValues("ping").Set(float64(metrics.Node.TypePing))
			clusterNodes.WithLabelValues("remote").Set(float64(metrics.Node.TypeRemote))
			clusterNodes.WithLabelValues("unknown").Set(float64(metrics.Node.TypeUnknown))

			// TODO: rename this metric with Total etc.
			//	clusterResourcesTotal.WithLabelValues("stopped").Add(float64(metrics.Resource.Stopped))
			//  clusterResources.WithLabelValues("started").Add(float64(metrics.Resource.Started))
			//	clusterResources.WithLabelValues("slave").Add(float64(metrics.Resource.Slave))
			//	clusterResources.WithLabelValues("master").Add(float64(metrics.Resource.Master))

			// this will produce a metric like this:
			// cluster_resources{node="dma-dog-hana01" resource_name="RA1"  role="master"} 1
			for _, nod := range status.Nodes.Node {
				for _, rsc := range nod.Resources {
					// TODO: FIXME FIND a mechanism to count the resources:
					// gauge2, err := pipelineCountMetric.GetMetricWithLabelValues("pipeline2")
					clusterResources.WithLabelValues(nod.Name, rsc.ID, rsc.Role).Set(float64(1))
				}
			}

			// TODO: this is historically, we might don't need to do like this. investigate on this later
			keys := make([]string, len(metrics.PerNode))
			i := 0
			for k := range metrics.PerNode {
				keys[i] = k
				i++
			}
			sort.Strings(keys)
			for _, k := range keys {
				node := metrics.PerNode[k]
				clusterResourcesRunning.WithLabelValues(k).Set(float64(node.ResourcesRunning))

			}
			// TODO: make this configurable later
			time.Sleep(2 * time.Second)

		}
	}()

	fmt.Println("[INFO]: Serving metrics on port", *portNumber)
	log.Fatal(http.ListenAndServe(*portNumber, nil))
}
