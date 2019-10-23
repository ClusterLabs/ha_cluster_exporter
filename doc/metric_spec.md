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

0. [Sample](../test/pacemaker.metrics)
1. [`ha_cluster_pacemaker_nodes`](#ha_cluster_pacemaker_nodes)
2. [`ha_cluster_pacemaker_nodes_total`](#ha_cluster_pacemaker_nodes_total)
3. [`ha_cluster_pacemaker_resources`](#ha_cluster_pacemaker_resources)
4. [`ha_cluster_pacemaker_resources_total`](#ha_cluster_pacemaker_resources_total)


### `ha_cluster_pacemaker_nodes`

#### Description

The nodes in the cluster; one line per `name`, per `status`.  
Either the value is `1`, or the line is absent altogether.

#### Labels

- `name`: name of the node (usually the hostname).
- `status`: one of `online|standby|standby_onfail|maintanance|pending|unclean|shutdown|expected_up|dc`. 
- `type`: one of `member|ping|remote`.

The total number of lines for this metric will be the cardinality of `name` times the cardinality of `status`.


### `ha_cluster_pacemaker_nodes_total` 

#### Description

The total number of *configured* nodes in the cluster. This value is mostly static and *does not* take into account the status of the nodes. It only changes when the Pacemaker configuration changes.


### `ha_cluster_pacemaker_resources` 

#### Description

The resources in the cluster; one line per `id`, per `status`.  
Either the value is `1`, or the line is absent altogether.

#### Labels

- `id`: the unique resource name.
- `node`: the name of the node hosting the resource.
- `managed`: either `true` or `false`.
- `role`:  one of `started|stopped|master|slave` or one of `starting|stopping|migrating|promoting|demoting`.
- `status` one of `active|orphaned|blocked|failed|failure_ignored`.

The total number of lines for this metric will be the cardinality of `id` times the cardinality of `status`.


### `ha_cluster_pacemaker_resources_total` 

#### Description

The total number of *configured* resources in the cluster. This value is mostly static and *does not* take into account the status of the resources. It only changes when the Pacemaker configuration changes.


## Corosync

The Corosync subsystem collects cluster quorum votes and ring status by parsing the output of `corosync-quorumtool` and `corosync-cfgtool`.

0. [Sample](../test/corosync.metrics)
1. [`ha_cluster_corosync_quorate`](#ha_cluster_corosync_quorate)
2. [`ha_cluster_corosync_quorum_votes`](#ha_cluster_corosync_quorum_votes)
3. [`ha_cluster_corosync_ring_errors_total`](#ha_cluster_corosync_ring_errors_total)


### `ha_cluster_corosync_quorate`

#### Description

Whether or not the cluster is quorate.  
Value is either `1` or `0`.


### `ha_cluster_corosync_quorum_votes`

#### Description

Cluster quorum votes; one line per type.

#### Labels

- `type`: one of `expected_votes|highest_expected|total_votes|quorum`


### `ha_cluster_corosync_ring_errors_total`

#### Description

Total number of corosync ring errors.


## SBD

The SBD subsystems collect devices stats by parsing its configuration the output of `sbd --dump`.

0. [Sample](../test/sbd.metrics)
1. [`ha_cluster_sbd_device_status`](#ha_cluster_sbd_device_status)

### `ha_cluster_sbd_device_status`

#### Description

Whether or not an SBD device is healthy. One line per `device`.  
Value is either `1` or `0`.

#### Labels

- `device`: the path of the device.

The total number of lines for this metric will be the cardinality of `device`.


## DRBD

The DRBD subsystems collect devices stats by parsing its configuration the JSON output of `drbdsetup`.

0. [Sample](../test/drbd.metrics)
1. [`ha_cluster_drbd_resources`](#ha_cluster_drbd_resources)
2. [`ha_cluster_drbd_connections`](#ha_cluster_drbd_connections)


### `ha_cluster_drbd_connections`

#### Description

The DRBD resource connections; 1 line per per `resource`, per `peer_node_id`  
Either the value is `1`, or the line is absent altogether.

#### Labels

- `resource`: the resource this connection is for.
- `peer_node_id`: the id of the node this connection is for
- `peer_role`: one of `primary|secondary|unknown`
- `volume`: the volume number
- `peer_disk_state`: one of `attaching|failed|negotiating|inconsistent|outdated|dunknown|consistent|uptodate`

The total number of lines for this metric will be the cardinality of `resource` times the cardinality of `peer_node_id`.


### `ha_cluster_drbd_resources`

#### Description

The DRBD resources; 1 line per `name`, per `volume`  
Either the value is `1`, or the line is absent altogether.

#### Labels

- `name`: the name of the resource.
- `role`: one of `primary|secondary|unknown`
- `volume`: the volume number
- `disk_state`: one of `attaching|failed|negotiating|inconsistent|outdated|dunknown|consistent|uptodate`

The total number of lines for this metric will be the cardinality of `name` times the cardinality of `volume`.
