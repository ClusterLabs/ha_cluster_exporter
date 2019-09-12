package main

import (
	"encoding/xml"
	"flag"
	"log"
	"net/http"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type crmMon struct {
	Version string  `xml:"version,attr"`
	Summary summary `xml:"summary"`
	Nodes   nodes   `xml:"nodes"`
}

type summary struct {
	Nodes     nodesConfigured     `xml:"nodes_configured"`
	Resources resourcesConfigured `xml:"resources_configured"`
}

type nodesConfigured struct {
	Number int `xml:"number,attr"`
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
	Node     nodeMetrics
	Resource resourceMetrics
	PerNode  map[string]perNodeMetrics
}

type nodeMetrics struct {
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

type resourceMetrics struct {
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

type perNodeMetrics struct {
	ResourcesRunning int
}

func parseMetrics(status *crmMon) *clusterMetrics {
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
	clusterNodeConf = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_nodes_configured",
		Help: "Number of nodes configured in ha cluster",
	})

	// a gauge metric with label
	clusterNodes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_nodes",
			Help: "cluster nodes metrics",
		}, []string{"type", "method"})
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(clusterNodeConf)
	prometheus.MustRegister(clusterNodes)
}

var addr = flag.String("listen-address", ":9002", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()
	// get cluster status xml
	monxml, err := exec.Command("/usr/sbin/crm_mon", "-1", "--as-xml", "--group-by-node", "--inactive").Output()
	if err != nil {
		panic(err)
	}

	var status crmMon
	err = xml.Unmarshal(monxml, &status)
	if err != nil {
		panic(err)
	}

	metrics := parseMetrics(&status)
	// add metrics
	clusterNodeConf.Set(float64(metrics.Node.Configured))
	clusterNodes.WithLabelValues("type", "member").Add(float64(metrics.Node.TypeMember))
	clusterNodes.WithLabelValues("type", "ping").Add(float64(metrics.Node.TypePing))
	clusterNodes.WithLabelValues("type", "remote").Add(float64(metrics.Node.TypeRemote))
	clusterNodes.WithLabelValues("type", "unknown").Add(float64(metrics.Node.TypeUnknown))

	// serve metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}

// TODO: implement this

// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_configured %v\n", metrics.Node.Configured))
// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_online %v\n", metrics.Node.Online))
// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_standby %v\n", metrics.Node.Standby))
// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_standby_onfail %v\n", metrics.Node.StandbyOnFail))
// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_maintenance %v\n", metrics.Node.Maintenance))
// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_pending %v\n", metrics.Node.Pending))
// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_unclean %v\n", metrics.Node.Unclean))
// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_shutdown %v\n", metrics.Node.Shutdown))
// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_expected_up %v\n", metrics.Node.ExpectedUp))
// 	io.WriteString(w, fmt.Sprintf("cluster_nodes_dc %v\n", metrics.Node.DC))

// 	// sort the keys to get consistent output
// 	keys := make([]string, len(metrics.PerNode))
// 	i := 0
// 	for k := range metrics.PerNode {
// 		keys[i] = k
// 		i++
// 	}
// 	sort.Strings(keys)
// 	for _, k := range keys {
// 		node := metrics.PerNode[k]
// 		io.WriteString(w, fmt.Sprintf("cluster_resources_running{node=\"%v\"} %v\n", k, node.ResourcesRunning))
// 	}
// 	io.WriteString(w, fmt.Sprintf("cluster_resources_configured %v\n", metrics.Resource.Configured))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources_unique %v\n", metrics.Resource.Unique))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources_disabled %v\n", metrics.Resource.Disabled))
// 	//io.WriteString(w, fmt.Sprintf("cluster_resources{role=\"stopped\"} %v\n", metrics.Resource.Stopped))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources{role=\"started\"} %v\n", metrics.Resource.Started))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources{role=\"slave\"} %v\n", metrics.Resource.Slave))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources{role=\"master\"} %v\n", metrics.Resource.Master))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources_active %v\n", metrics.Resource.Active))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources_orphaned %v\n", metrics.Resource.Orphaned))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources_blocked %v\n", metrics.Resource.Blocked))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources_managed %v\n", metrics.Resource.Managed))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources_failed %v\n", metrics.Resource.Failed))
// 	io.WriteString(w, fmt.Sprintf("cluster_resources_failure_ignored %v\n", metrics.Resource.FailureIgnored))
