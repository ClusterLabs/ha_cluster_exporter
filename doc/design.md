# Design Notes

This document describes the rationale behind design decisions takend during the development of this project.

## Goals

- Export runtime statistics about the various HA cluster components from existing data sources, to be consumed in a Prometheus monitoring stack.

## Non-goals

- Maintain an internal, consistent, persisting representation of the cluster state; since the original source of truth is distributed, we want to avoid the complexity of a stateful middleware.


## Structure

The project consist in a small HTTP application that exposes runtime data in a line protocol.

A series of `prometheus.Collector` implementations, one for each cluster component (that we'll call _subsystems_) are instantiated in the main application entry point, registered with the Prometheus client, and then exposed via its HTTP handler.

Each collector `Collect` method will be called concurrently by the client itself in an internal worker goroutine.

The sources hence are read every time an HTTP request comes, and the collected data is not shared: its lifecycle corresponds with the request's.

To avoid concurrent reads of the same source, all `Collect` methods are serialized with a mutex.

## Collectors

The collectors are very simple: they usually just invoke a bunch of system commands, then parse the output into bespoke data structures that can be used to build Prometheus metrics.

More details about these metrics can be found in the [metrics specification document](metrics.md).
