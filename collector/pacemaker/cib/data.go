package cib

/*
The Cluster Information Base (Root) is an XML representation of the cluster’s configuration and the state of all nodes and resources.
The Root manager (pacemaker-based) keeps the Root synchronized across the cluster, and handles requests to modify it.

https://clusterlabs.org/pacemaker/doc/en-US/Pacemaker/2.0/html-single/Pacemaker_Administration/index.html

*/

type Root struct {
	Configuration struct {
		Nodes []struct {
			Id                 string      `xml:"id,attr"`
			Uname              string      `xml:"uname,attr"`
			InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
		} `xml:"nodes>node"`
		Resources struct {
			Primitives []Primitive `xml:"primitive"`
			Masters    []Clone     `xml:"master"`
			Clones     []Clone     `xml:"clone"`
		} `xml:"resources"`
		Constraints struct {
			RscLocations []struct {
				Id       string `xml:"id,attr"`
				Node     string `xml:"node,attr"`
				Resource string `xml:"rsc,attr"`
				Role     string `xml:"role,attr"`
				Score    string `xml:"score,attr"`
			} `xml:"rsc_location"`
		} `xml:"constraints"`
	} `xml:"configuration"`
}

type Attribute struct {
	Id    string `xml:"id,attr"`
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Primitive struct {
	Id                 string      `xml:"id,attr"`
	Class              string      `xml:"class,attr"`
	Type               string      `xml:"type,attr"`
	Provider           string      `xml:"provider,attr"`
	InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
	MetaAttributes     []Attribute `xml:"meta_attributes>nvpair"`
	Operations         []struct {
		Id       string `xml:"id,attr"`
		Name     string `xml:"name,attr"`
		Role     string `xml:"role,attr"`
		Interval int    `xml:"interval,attr"`
		Timeout  int    `xml:"timeout,attr"`
	} `xml:"operations>op"`
}

type Clone struct {
	Id             string      `xml:"id,attr"`
	MetaAttributes []Attribute `xml:"meta_attributes>nvpair"`
	Primitive      Primitive   `xml:"primitive"`
}
