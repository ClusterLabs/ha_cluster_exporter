# ha_cluster_exporter

This prometheus exporter is used to serve metrics for pacemaker https://github.com/ClusterLabs/pacemaker

It should run inside a node of the cluster.

# Usage

Build the binary with `make` and run it, it will expose the metrics on port `9001` by default.

Use `ha_cluster_exporter -h` for options.

# Devel:

Build the exporter and copy it to a node of the cluster
