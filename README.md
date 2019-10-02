# ha_cluster_exporter

[![Build Status](https://travis-ci.org/ClusterLabs/ha_cluster_exporter.svg?branch=master)](https://travis-ci.org/ClusterLabs/ha_cluster_exporter)


This prometheus exporter is used to serve metrics for pacemaker https://github.com/ClusterLabs/pacemaker

It should run inside a node of the cluster or both.

# Usage:

You can find RPM pkgs for the exporter here: https://build.opensuse.org/package/show/server:monitoring/prometheus-ha_cluster_exporter

Once installed run **inside a cluster node** the exporter with: 

`systemctl start prometheus-ha_cluster_exporter`, by default it will expose on `http://YOUR_HOST_IP:9002/metrics`.

If you open a web-browser it will serve the metrics. 

The exporter can't work outside a HA cluster node

**Hint:**
For a terraform deployment you can read also : https://github.com/SUSE/ha-sap-terraform-deployments

# Features:

- expose cluster node and resource metrics via `crm_mon` (pacemaker data xml)

- expose corosync metrics (ring errors, quorum metrics)

- expose SBD disk health metrics

# Devel:

Build the binary with `make` and run it inside a node of the ha cluster , it will expose the metrics on port `9002` by default.

Use `ha_cluster_exporter -h` for options.

#### Design:

For the technical design of the exporter have look at [design](doc/design.md) (this is focused on cluster_metrics)

