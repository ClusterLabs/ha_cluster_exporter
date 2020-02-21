# Design Notes

This document describes the rationale behind design decisions taken during the development of this project.

## Goals

- Export runtime statistics about the various ClusterLabs Linux HA cluster components from existing data sources, to be consumed by a Prometheus monitoring stack.

## Non-goals

- Maintain an internal, consistent, persisting representation of the cluster state; since the original source of truth is distributed, we want to avoid the complexity of a stateful middleware.


## Structure

The project consist in a small HTTP application that exposes runtime data in a line protocol.
  
A series of "metric collectors" are consumed by the main application entry point, `ha_cluster_exporter.go`, where they are registered with the Prometheus client and then exposed via its HTTP handler.

Concurrency is handled internally by a worker pool provided by the Prometheus library, but this implementation detail is completely obfuscated to the consumers.

The data sources are read every time an HTTP request comes, and the collected metrics are not shared: their lifecycle corresponds with the request's.

The `internal` package contains common code shared among all the other packages, but not intended for usage outside this projects.

## Collectors

Inside the `collector` package, you wil find the code of the main logic of the project: these are a number of [`prometheus.Collector`](https://github.com/prometheus/client_golang/blob/b25ce2693a6de99c3ea1a1471cd8f873301a452f/prometheus/collector.go#L16-L63) implementations, one for each cluster component (that we'll call _subsystems_), like Pacemaker, or Corosync.

Common functionality is provided by composing the [`DefaultCollector`](../collector/default_collector.go). 

Each subsystem collector has a dedicated package; some are very simple, some are little more nuanced. In general, they depend on external, globally available, system tools, to introspect the subsystems. 

The collectors usually just invoke these system commands, parsing the output into bespoke data structures.
When building these data structures involves a significant amount of code, for a better separation of concerns this responsibility is extracted in dedicated subpackages, like [`collector/pacemaker/cib`](../collector/pacemaker/cib).

The data structures are then used by the collectors to build the Prometheus metrics. 

More details about the metrics themselves can be found in the [metrics](metrics.md) document.
