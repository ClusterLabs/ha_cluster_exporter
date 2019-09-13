package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sort"

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

	clusterResourcesConf = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_configured",
		Help: "Number of configured resources in ha cluster",
	})

	clusterResourcesUnique = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_unique",
		Help: "Number of uniques resources in ha cluster",
	})

	clusterResourcesDisabled = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_disabled",
		Help: "Number resources disabled in ha cluster",
	})

	clusterResourcesActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_active",
		Help: "Number resources active in ha cluster",
	})

	clusterResourcesOrphaned = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_orphaned",
		Help: "Number resources orphaned in ha cluster",
	})

	clusterResourcesBlocked = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_blocked",
		Help: "Number resources blocked in ha cluster",
	})

	clusterResourcesManaged = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_managed",
		Help: "Number resources managed in ha cluster",
	})

	clusterResourcesFailed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_failed",
		Help: "Number resources failed in ha cluster",
	})

	clusterResourcesFailedIgnored = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_resources_failed_ignored",
		Help: "Number resources failure ignored in ha cluster",
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
		}, []string{"role"})
)

func init() {
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
	prometheus.MustRegister(clusterResourcesConf)
	prometheus.MustRegister(clusterResourcesUnique)
	prometheus.MustRegister(clusterResourcesDisabled)
	prometheus.MustRegister(clusterResourcesActive)
	prometheus.MustRegister(clusterResourcesOrphaned)
	prometheus.MustRegister(clusterResourcesBlocked)
	prometheus.MustRegister(clusterResourcesManaged)
	prometheus.MustRegister(clusterResourcesFailed)
	prometheus.MustRegister(clusterResourcesFailedIgnored)

	// metrics with labels
	prometheus.MustRegister(clusterNodes)
	prometheus.MustRegister(clusterResourcesRunning)
	prometheus.MustRegister(clusterResources)

}

var portNumber = flag.String("port", ":9001", "The port number to listen on for HTTP requests.")

func main() {
	flag.Parse()
	// get cluster status xml
	monxml, err := exec.Command("/usr/sbin/crm_mon", "-1", "--as-xml", "--group-by-node", "--inactive").Output()
	if err != nil {
		fmt.Println("[ERROR]: crm_mon command was not executed correctly. Did you have crm_mon installed ?")
		panic(err)
	}

	var status crmMon
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
	// add genric resource metrics
	clusterResourcesUnique.Set(float64(metrics.Resource.Unique))
	clusterResourcesDisabled.Set(float64(metrics.Resource.Disabled))
	clusterResourcesConf.Set(float64(metrics.Resource.Configured))
	clusterResourcesActive.Set(float64(metrics.Resource.Active))
	clusterResourcesOrphaned.Set(float64(metrics.Resource.Orphaned))
	clusterResourcesBlocked.Set(float64(metrics.Resource.Blocked))
	clusterResourcesManaged.Set(float64(metrics.Resource.Managed))
	clusterResourcesFailed.Set(float64(metrics.Resource.Failed))
	clusterResourcesFailedIgnored.Set(float64(metrics.Resource.FailureIgnored))

	// metrics with labels
	clusterNodes.WithLabelValues("member").Add(float64(metrics.Node.TypeMember))
	clusterNodes.WithLabelValues("ping").Add(float64(metrics.Node.TypePing))
	clusterNodes.WithLabelValues("remote").Add(float64(metrics.Node.TypeRemote))
	clusterNodes.WithLabelValues("unknown").Add(float64(metrics.Node.TypeUnknown))

	clusterNodes.WithLabelValues("stopped").Add(float64(metrics.Resource.Stopped))
	clusterNodes.WithLabelValues("started").Add(float64(metrics.Resource.Started))
	clusterNodes.WithLabelValues("slave").Add(float64(metrics.Resource.Slave))
	clusterNodes.WithLabelValues("master").Add(float64(metrics.Resource.Master))

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
		clusterResourcesRunning.WithLabelValues(k).Add(float64(node.ResourcesRunning))
	}

	// serve metrics
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("[INFO]: Serving metrics on port", *portNumber)
	log.Fatal(http.ListenAndServe(*portNumber, nil))
}
