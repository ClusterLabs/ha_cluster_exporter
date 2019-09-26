# ha_cluster_exporter

This prometheus exporter is used to serve metrics for pacemaker https://github.com/ClusterLabs/pacemaker

It should run inside a node of the cluster or both.

# Features:

- expose cluster node and resource metrics via `crm_mon` (pacemaker data xml)

- expose corosync metrics **not done yet WIP**

# Design:

For the technical design of the exporter have look at [design](doc/design.md) (this is focused on cluster_metrics)

# Devel:

Build the binary with `make` and run it inside a node of the ha cluster , it will expose the metrics on port `9001` by default.

Use `ha_cluster_exporter -h` for options.

# Packages:

You can find Packages here: https://build.opensuse.org/package/show/server:monitoring/prometheus-ha_cluster_exporter
