# Metrics specification

This document describes the metrics exposed by `ha_cluster_exporter`.

General notes:
- All the metrics are _namespaced_ with the prefix `ha_cluster`, which is followed by a _subsystem_, and both are in turn composed into a _Fully Qualified Name_ (FQN) of each metrics.
- All the metrics and labels _names_ are in snake_case, as conventional with Prometheus. That said, as much as we'll try to keep this consistent throughout the project, the label _values_ may not actually follow this convention, though (e.g. value is a hostname).

- If the `enable-timestamps` option is on, all the metrics will be timestamped with the Unix epoch time in milliseconds.

These are the currently implemented subsystems.

1. [Pacemaker](#pacemaker)
2. [Corosync](#corosync)
3. [SBD](#sbd)
4. [DRBD](#drbd)
5. [Scrape](#scrape)


## Pacemaker 

The Pacemaker subsystem collects an atomic snapshot of the HA cluster directly from the XML CIB of Pacemaker via `crm_mon`.

0. [Sample](../test/pacemaker.metrics)
1. [`ha_cluster_pacemaker_config_last_change`](#ha_cluster_pacemaker_config_last_change)
2. [`ha_cluster_pacemaker_fail_count`](#ha_cluster_pacemaker_fail_count)
3. [`ha_cluster_pacemaker_location_constraints`](#ha_cluster_pacemaker_location_constraints)
4. [`ha_cluster_pacemaker_migration_threshold`](#ha_cluster_pacemaker_migration_threshold)
5. [`ha_cluster_pacemaker_nodes`](#ha_cluster_pacemaker_nodes)
6. [`ha_cluster_pacemaker_node_attributes`](#ha_cluster_pacemaker_node_attributes)
7. [`ha_cluster_pacemaker_resources`](#ha_cluster_pacemaker_resources)
8. [`ha_cluster_pacemaker_stonith_enabled`](#ha_cluster_pacemaker_stonith_enabled)


### `ha_cluster_pacemaker_config_last_change`

#### Description

The value of this metric is a Unix timestamp in seconds, converted to a float, corresponding to the last time Pacemaker configuration changed.
The metric is in turn timestamped with the time it was last checked.


### `ha_cluster_pacemaker_fail_count`

#### Description

The number of fail count per node and resource ID.  
The value is an integer ranging from 0 to `+Inf`.    
The actual maximum integer value depends on Pacemaker internals, so please refer to upstream documentation for further information.


### `ha_cluster_pacemaker_location_constraints`

#### Description

Resource location constraints.  
The value of the metric is the **score** of the constraint, represented by an integer ranging from `-Inf` to `+Inf`.  
The actual minimum and maximum integer values depend on Pacemaker internals, so please refer to upstream documentation for further information.

#### Labels

- `constraint`: the unique string identifier of the constraint.
- `node`: the node the constraint applies to.
- `resource`: the resource the constraint applies to.
- `role`: the resource role the constraint applies to, if any.


### `ha_cluster_pacemaker_migration_threshold`

#### Description

The number of migration threshold pro node and resource ID set by a pacemaker cluster. 
Possible values are positive numbers.


### `ha_cluster_pacemaker_nodes`

#### Description

The status of each node in the cluster; it will have one line for each possible `status` of each `node`.
A value of `1` means the node is in the status specified by the `status` label, a value of `0` means it is not.

#### Labels

- `node`: name of the node (usually the hostname).
- `status`: one of `online|standby|standby_onfail|maintanance|pending|unclean|shutdown|expected_up|dc`. 
- `type`: one of `member|ping|remote`.


### `ha_cluster_pacemaker_node_attributes`

#### Description

This metric exposes in its labels raw, opaque, cluster metadata, called node attributes, which often leveraged by Resource Agents.  
The value of each line will always be `1`.

#### Labels

- `node`: name of the node (usually the hostname).
- `name`: name of the attribute.
- `value`: value of the attribute.


### `ha_cluster_pacemaker_resources` 

#### Description

The status of each resource in the cluster; it will have one line for each possible `status` of each `resource`.  
A value of `1` means the resource is in the status specified by the `status` label, a value of `0` means it is not.

#### Labels

- `agent`: the name of the resource agent for this resource.
- `clone`: the name of the clone this resource belongs to, if any.
- `group`: the name of the group this resource belongs to, if any.
- `managed`: either `true` or `false`.
- `node`: the name of the node hosting the resource.
- `resource`: the unique resource name.
- `role`:  one of `started|stopped|master|slave` or one of `starting|stopping|migrating|promoting|demoting`.
- `status`: one of `active|orphaned|blocked|failed|failure_ignored`.


### `ha_cluster_pacemaker_stonith_enabled`

#### Description

Whether or not stonith is enabled in the cluster.  
Value is either `1` or `0`.


## Corosync

The Corosync subsystem collects cluster quorum votes and ring status by parsing the output of `corosync-quorumtool` and `corosync-cfgtool`.

0. [Sample](../test/corosync.metrics)
1. [`ha_cluster_corosync_member_votes`](#ha_cluster_corosync_member_votes)
2. [`ha_cluster_corosync_quorate`](#ha_cluster_corosync_quorate)
3. [`ha_cluster_corosync_quorum_votes`](#ha_cluster_corosync_quorum_votes)
4. [`ha_cluster_corosync_ring_errors`](#ha_cluster_corosync_ring_errors)
5. [`ha_cluster_corosync_rings`](#ha_cluster_corosync_rings)


### `ha_cluster_corosync_member_votes`

#### Description

How many votes each member node has contributed with to the current quorum

#### Labels

- `node_id`: the internal corosync identifier associated to this node.
- `node`: the name of the node; usually the hostname.
- `local`: whether or not this is the local node.


### `ha_cluster_corosync_quorate`

#### Description

Whether or not the cluster is quorate.  
Value is either `1` or `0`.


### `ha_cluster_corosync_quorum_votes`

#### Description

Cluster quorum votes; one line per type.

#### Labels

- `type`: one of `expected_votes|highest_expected|total_votes|quorum`


### `ha_cluster_corosync_ring_errors`

#### Description

The total number of faulty corosync rings.


### `ha_cluster_corosync_rings`

#### Description

The status of each Corosync ring; `1` means healthy, `0` means faulty.

#### Labels

- `ring_id`: the internal Corosync ring identifier; usually corresponds to the first member node to join.  
- `node_id`: the internal Corosync identifier of the local node.
- `number`: the ring number.
- `address`: the IP address locally linked to this ring.


## SBD

The SBD subsystems collect devices stats by parsing its configuration and the output of `sbd --dump`.

0. [Sample](../test/sbd.metrics)
1. [`ha_cluster_sbd_devices`](#ha_cluster_sbd_devices)
2. [`ha_cluster_sbd_timeouts`](#ha_cluster_sbd_timeouts)

### `ha_cluster_sbd_devices`

#### Description

The SBD devices in the cluster; one line per device.  
Either the value is `1`, or the line is absent altogether.

#### Labels

- `device`: the path of the SBD device
- `status`: one of `healthy|unhealthy`

The total number of lines for this metric will be the cardinality of `device`.

### `ha_cluster_sbd_timeouts`

#### Description

The SBD timeouts pro SBD device
Value is an integer expessing the timeout

#### Labels

- `device`: the path of the SBD device
- `type`:  either `watchdog` or `msgwait`


## DRBD

The DRBD subsystems collect devices stats by parsing its configuration the JSON output of `drbdsetup`.

0. [Sample](../test/drbd.metrics)
1. [`ha_cluster_drbd_resources`](#ha_cluster_drbd_resources)
2. [`ha_cluster_drbd_written`](#ha_cluster_drbd_written)
3. [`ha_cluster_drbd_read`](#ha_cluster_drbd_read)
4. [`ha_cluster_drbd_al_writes`](#ha_cluster_al_writes)
5. [`ha_cluster_drbd_bm_writes`](#ha_cluster_bm_writes)
6. [`ha_cluster_drbd_upper_pending`](#ha_cluster_drbd_upper_pending)
7. [`ha_cluster_drbd_lower_pending`](#ha_cluster_drbd_lower_pending)
8. [`ha_cluster_drbd_quorum`](#ha_cluster_drbd_quorum)
9. [`ha_cluster_drbd_connections`](#ha_cluster_drbd_connections)
10. [`ha_cluster_drbd_connections_sync`](#ha_cluster_drbd_connections_sync)
11. [`ha_cluster_drbd_connections_received`](#ha_cluster_drbd_connections_received)
12. [`ha_cluster_drbd_connections_sent`](#ha_cluster_drbd_connections_sent)
13. [`ha_cluster_drbd_connections_pending`](#ha_cluster_drbd_connections_pending)
14. [`ha_cluster_drbd_connections_unacked`](#ha_cluster_drbd_connections_unacked)
15. [`ha_cluster_drbd_split_brain`](#ha_cluster_drbd_split_brain)

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

### `ha_cluster_drbd_connections_sync`

#### Description

The DRBD disk connections in sync percentage. Values are float from `0` to `100.00`.

#### Labels

- `resource`: the resource this connection is for.
- `peer_node_id`: the id of the node this connection is for
- `volume`: the volume number

### `ha_cluster_drbd_connections_received`

#### Description

Volume of net data received from the partner via the network connection in KiB; 1 line per per `resource`, per `peer_node_id`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the resource this connection is for.
- `peer_node_id`: the id of the node this connection is for
- `volume`: the volume number

### `ha_cluster_drbd_connections_sent`

#### Description

Volume of net data sent to the partner via the network connection in KiB; 1 line per per `resource`, per `peer_node_id`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the resource this connection is for.
- `peer_node_id`: the id of the node this connection is for
- `volume`: the volume number

### `ha_cluster_drbd_connections_pending`

#### Description

Number of requests sent to the partner that have not yet been received; 1 line per per `resource`, per `peer_node_id`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the resource this connection is for.
- `peer_node_id`: the id of the node this connection is for
- `volume`: the volume number

### `ha_cluster_drbd_connections_unacked`

#### Description

Number of requests received by the partner but have not yet been acknowledged; 1 line per per `resource`, per `peer_node_id`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the resource this connection is for.
- `peer_node_id`: the id of the node this connection is for
- `volume`: the volume number

### `ha_cluster_drbd_resources`

#### Description

The DRBD resources; 1 line per `name`, per `volume`  
Either the value is `1`, or the line is absent altogether.

#### Labels

- `resource`: the name of the resource.
- `role`: one of `primary|secondary|unknown`
- `volume`: the volume number
- `disk_state`: one of `attaching|failed|negotiating|inconsistent|outdated|dunknown|consistent|uptodate`

The total number of lines for this metric will be the cardinality of `name` times the cardinality of `volume`.

### `ha_cluster_drbd_written`

#### Description

Amount in KiB written to the DRBD resource; 1 line per `resource`, per `volume`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the name of the resource.
- `volume`: the volume number

### `ha_cluster_drbd_read`

#### Description

Amount in KiB read from the DRBD resource; 1 line per `resource`, per `volume`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the name of the resource.
- `volume`: the volume number

### `ha_cluster_drbd_al_writes`

#### Description

Number of updates of the activity log area of the meta data; 1 line per `resource`, per `volume`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the name of the resource.
- `volume`: the volume number

### `ha_cluster_drbd_bm_writes`

#### Description

Number of updates of the bitmap area of the meta data; 1 line per `resource`, per `volume`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the name of the resource.
- `volume`: the volume number

### `ha_cluster_drbd_upper_pending`

#### Description

Number of block I/O requests forwarded to DRBD, but not yet answered by DRBD; 1 line per `resource`, per `volume`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the name of the resource.
- `volume`: the volume number

### `ha_cluster_drbd_lower_pending`

#### Description

Number of open requests to the local I/O sub-system issued by DRBD; 1 line per `resource`, per `volume`
Value is an integer greater than or equal to `0`.

#### Labels

- `resource`: the name of the resource.
- `volume`: the volume number

### `ha_cluster_drbd_quorum`

#### Description

Quorum status of the DRBD resource according to it's configured quorum policies; 1 line per `resource`, per `volume`
Value is `1` when quorate, or `0` when inquorate.

#### Labels

- `resource`: the name of the resource.
- `volume`: the volume number

### `ha_cluster_drbd_split_brain`

#### Description

This metric signal if there is a split brain occurring per resource and volume.
Either the value is `1`, or the line is absent altogether.

This metric is a special metric compared to others, because in order to make this metric work you will need to setup a DRBD custom split-brain handler. Look at the end.

#### Labels

- `resource`: the name of the resource.
- `volume`: the volume number

#### Setting up the DRBD split-brain hook

In order to get the `split_brain` metric working:

1) copy hook into all drbd nodes:

get the hook from:
https://github.com/SUSE/ha-sap-terraform-deployments/blob/72c9d3ecf6c3f6dd18ccb7bcbde4b40722d5c641/salt/drbd_node/files/notify-split-brain-haclusterexporter-suse-metric.sh

2) on the DRBD configuration enable the hook:

```split_brain: "/usr/lib/drbd/notify-split-brain-haclusterexporter-suse-metric.sh"```

Refer to upstream doc: https://docs.linbit.com/docs/users-guide-8.4/#s-configure-split-brain-behavior

It is important for the exporter that he hook should create the files in that location and naming. 

Remember to remove the files manually after the split brain is solved


## Scrape

The `scrape` subsystem is a generic namespace dedicated to internal instrumentation of the exporter itself.

1. [`ha_cluster_scrape_duration_seconds`](#ha_cluster_scrape_duration_seconds)
2. [`ha_cluster_scrape_success`](#ha_cluster_scrape_success)

### `ha_cluster_scrape_duration_seconds`

The duration of a collector scrape in seconds. 

#### Labels

- `collector`: collector names correspond to the subsystem they collect metrics from.

#### Example

```
# TYPE ha_cluster_scrape_duration_seconds gauge
ha_cluster_scrape_duration_seconds{collector="pacemaker"} 1.234
```

### `ha_cluster_scrape_success`

Whether a collector succeeded. 

Collectors may gracefully fail, but this won't prevent them from continuing running. 

If some metrics could not be scraped, the value of this metric will be `0`.  
In such cases, you shall find more details in the exporter logs. 

#### Labels

- `collector`: collector names correspond to the subsystem they collect metrics from.

#### Examaple

```
# TYPE ha_cluster_scrape_success gauge
ha_cluster_scrape_success{collector="pacemaker"} 1
```
