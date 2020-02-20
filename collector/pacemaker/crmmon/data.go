package crmmon

// *** crm_mon XML unserialization structures

type Root struct {
	Version string `xml:"version,attr"`
	Summary struct {
		Nodes struct {
			Number int `xml:"number,attr"`
		} `xml:"nodes_configured"`
		LastChange struct {
			Time string `xml:"time,attr"`
		} `xml:"last_change"`
		Resources struct {
			Number   int `xml:"number,attr"`
			Disabled int `xml:"disabled,attr"`
			Blocked  int `xml:"blocked,attr"`
		} `xml:"resources_configured"`
		ClusterOptions struct {
			StonithEnabled bool `xml:"stonith-enabled,attr"`
		} `xml:"cluster_options"`
	} `xml:"summary"`
	Nodes       []Node `xml:"nodes>node"`
	NodeHistory struct {
		Node []struct {
			Name            string `xml:"name,attr"`
			ResourceHistory []struct {
				Name               string `xml:"id,attr"`
				MigrationThreshold int    `xml:"migration-threshold,attr"`
				FailCount          int    `xml:"fail-count,attr"`
			} `xml:"resource_history"`
		} `xml:"node"`
	} `xml:"node_history"`
}

type Node struct {
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
	Resources        []Resource `xml:"resource"`
}

type Resource struct {
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
