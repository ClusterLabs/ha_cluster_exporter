package internal

import "net/http"

func Landing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
<html>
<head>
	<title>ClusterLabs Linux HA Cluster Exporter</title>
</head>
<body>
	<h1>ClusterLabs Linux HA Cluster Exporter</h1>
	<h2>Prometheus exporter for Pacemaker based Linux HA clusters</h2>
	<ul>
		<li><a href="metrics">Metrics</a></li>
		<li><a href="https://github.com/ClusterLabs/ha_cluster_exporter" target="_blank">GitHub</a></li>
	</ul>
</body>
</html>
`))
}
