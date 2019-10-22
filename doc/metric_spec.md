# Metrics specification

This document describes the metrics exposed by `ha_cluster_exporter`.

General notes:
- All the metrics are _namespaced_ with the prefix `ha_cluster`, which is followed by a _subsystem_, and both are in turn composed into a _Fully Qualified Name_ (FQN) of each metrics.
- All the metrics and labels _names_ are in snake_case, as conventional with Prometheus. That said, as much as we'll try to keep this consistent throughout the project, the label _values_ may not actually follow this convention, though (e.g. value is a hostname).
- All the metrics are timestamped with the Unix epoch time in milliseconds; in the provided examples, this value will always be `1234`.
- Some metrics, like `ha_cluster_pacemaker_nodes`, `ha_cluster_pacemaker_resources`, share common traits:
  - their labels contain the relevant data you may want to track or use for aggregation and filtering;
  - either their value is `1`, or the line is absent altogether; this is because each line represents one entity of the cluster, but the exporter itself is stateless, i.e. we don't track the life-cycle of entities that do not exist anymore in the cluster.


These are the currently implemented subsystems.

1. [Pacemaker](#pacemaker)
2. [Corosync](#corosync)
3. [SBD](#sbd)
4. [DRBD](#drbd)


## Pacemaker 

The Pacemaker subsystem collects an atomic snapshot of the HA cluster directly from the XML CIB of Pacemaker via `crm_mon`.

1. [ha_cluster_pacemaker_nodes](#ha_cluster_pacemaker_nodes)
2. [ha_cluster_pacemaker_nodes_total](#ha_cluster_pacemaker_nodes_total)
3. [ha_cluster_pacemaker_resources](#ha_cluster_pacemaker_resources)
4. [ha_cluster_pacemaker_resources_total](#ha_cluster_pacemaker_resources_total)


### `ha_cluster_pacemaker_nodes`

#### Description

The nodes in the cluster; one line per `name`, per `status`.  
Either the value is `1`, or the line is absent altogether.

#### Labels

- `name`: name of the node (usually the hostname).
- `status`: one of `online|standby|standby_onfail|maintanance|pending|unclean|shutdown|expected_up|dc`. 
- `type`: one of `member|ping|remote`.

The total number of lines for this metric will be the cardinality of `name` times the cardinality of `status`.

#### Example

https://github.com/ClusterLabs/ha_cluster_exporter/blob/f4512578dc5bb6421a1813a378fff18acc27208d/test/pacemaker.metrics#L1-L7


### `ha_cluster_pacemaker_nodes_total` 

#### Description

The total number of *configured* nodes in the cluster. This value is mostly static and *does not* take into account the status of the nodes. It only changes when the Pacemaker configuration changes.

#### Example

https://github.com/ClusterLabs/ha_cluster_exporter/blob/f4512578dc5bb6421a1813a378fff18acc27208d/test/pacemaker.metrics#L8-L10


### `ha_cluster_pacemaker_resources` 

#### Description

The resources in the cluster; one line per `id`, per `status`.  
Either the value is `1`, or the line is absent altogether.

#### Labels

- `id`: the unique resource name.
- `node`: name of the node hosting the resource. 
- `managed`: either `true` or `false`.
- `role`:  one of `started|stopped|master|slave` or one of `starting|stopping|migrating|promoting|demoting`.
- `status` one of `active|orphaned|blocked|failed|failure_ignored`.

The total number of lines for this metric will be the cardinality of `id` times the cardinality of `status`.

#### Example

https://github.com/ClusterLabs/ha_cluster_exporter/blob/f4512578dc5bb6421a1813a378fff18acc27208d/test/pacemaker.metrics#L11-L18


### `ha_cluster_pacemaker_resources_total` 

#### Description

The total number of *configured* resources in the cluster. This value is mostly static and *does not* take into account the status of the resources. It only changes when the Pacemaker configuration changes.

#### Example

https://github.com/ClusterLabs/ha_cluster_exporter/blob/f4512578dc5bb6421a1813a378fff18acc27208d/test/pacemaker.metrics#L19-L21


## Corosync

The Corosync subsystem collects cluster quorum votes and ring status by parsing the output of `corosync-quorumtool` and `corosync-cfgtool`.

1. [ha_cluster_corosync_quorate](#ha_cluster_corosync_quorate)
2. [ha_cluster_corosync_quorum_votes](#ha_cluster_corosync_quorum_votes)
3. [ha_cluster_corosync_ring_errors_total](#ha_cluster_corosync_ring_errors_total)


### `ha_cluster_corosync_quorate`

#### Description

Whether or not the cluster is quorate.  
Value is either `1` or `0`.

#### Example

https://github.com/ClusterLabs/ha_cluster_exporter/blob/f4512578dc5bb6421a1813a378fff18acc27208d/test/corosync.metrics#L1-L3


### `ha_cluster_corosync_quorum_votes`

#### Description

Cluster quorum votes; one line per type.

#### Labels

- `type`: one of `expected_votes|highest_expected|total_votes|quorum`

#### Example

https://github.com/ClusterLabs/ha_cluster_exporter/blob/f4512578dc5bb6421a1813a378fff18acc27208d/test/corosync.metrics#L4-L9


### `ha_cluster_corosync_ring_errors_total`

#### Description

Total number of corosync ring errors.

#### Example

https://github.com/ClusterLabs/ha_cluster_exporter/blob/f4512578dc5bb6421a1813a378fff18acc27208d/test/corosync.metrics#L10-L2


## SBD

The SBD subsystems collect devices stats by parsing its configuration and parsing the output of `sbd --dump`.


### `ha_cluster_sbd_device_status`

#### Description

Health status for each SBD device; 1 is healthy, 0 is not.

#### Example

https://github.com/ClusterLabs/ha_cluster_exporter/blob/f4512578dc5bb6421a1813a378fff18acc27208d/test/sbd.metrics#L1-L4


## DRBD

`TODO`
