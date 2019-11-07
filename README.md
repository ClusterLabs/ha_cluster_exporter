# ha_cluster_exporter

[![Build Status](https://travis-ci.org/ClusterLabs/ha_cluster_exporter.svg?branch=master)](https://travis-ci.org/ClusterLabs/ha_cluster_exporter)


This is a Prometheus exporter used to export metrics of the various components of HA clusters.  

## Table of Contents
1. [Usage](#Usage)
2. [Features](#Features)
3. [Devel](#Devel)
4. [Design](#Design)

## Usage

You can find the RPM pkgs for the exporter here: https://build.opensuse.org/package/show/server:monitoring/prometheus-ha_cluster_exporter.

Once installed run the exporter **inside a cluster node** with: 

`systemctl start prometheus-ha_cluster_exporter`. By default it will show on `http://YOUR_HOST_IP:9002/metrics`.

If you open a web-browser it will serve the metrics. 
The exporter can't work outside a HA cluster node.

**Hint:**
For a terraform deployment you can also read: https://github.com/SUSE/ha-sap-terraform-deployments

## Features
It should be run inside one or more nodes composing the cluster. Running in every nodes is advised.

- show cluster node and resource metrics via `crm_mon` (pacemaker data xml)
- show corosync metrics (ring errors, quorum metrics)
- show SBD disk health metrics
- show DRBD metrics (local and remote disks resource metrics)

A complete list of the metrics can be found in the [metrics](doc/metric_spec.md) document.

## Devel

Build the binary with `make` and run it inside a node of the ha cluster, it will show the metrics on port `9002` by default.
Use `ha_cluster_exporter -h` for options.

#### Design

For the technical design of the exporter have look at [design](doc/design.md) (this is focused on cluster_metrics).

