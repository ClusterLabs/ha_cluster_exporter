# ha_cluster_exporter

This prometheus exporter is used to serve metrics for pacemaker https://github.com/ClusterLabs/pacemaker

It should run inside a node of the cluster or both.

# Design:

[design](doc/design.md)

# Usage

Build the binary with `make` and run it, it will expose the metrics on port `9001` by default.

Use `ha_cluster_exporter -h` for options.

# Devel:

Build the exporter and copy it to a node of the cluster

# Packages:

You can find a package for #openSUSE distro here: 

https://build.opensuse.org/package/show/network:ha-clustering:Factory/prometheus-ha_cluster_exporter

In this repo you can find also all the HA pkgs for openSUSE for having/building a cluster. See https://en.opensuse.org/openSUSE:High_Availability
