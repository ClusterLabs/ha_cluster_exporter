# Metrics specification:

This is a specification of metrics exposed by the ha_cluster exporter.

All metrics from the exporter start with the prefix `ha_cluster`

Below you have a complete specification, ordered by component.

1. [pacemaker](#pacemaker)
2. [drbd](#drbd)
3. [sbd](#sbd)
4. [corosyncl](#corosync)

# Pacemaker 

The Pacemaker cluster metrics are atomic metrics and represent and updated snapshot of the HA cluster, retrieved fetching the XML CIB of pacemaker.

Some of the pacemaker metrics like `ha_cluster_node_resources` and `ha_cluster_nodes` metrics with labels share a common trait:
 
they can be either set to `1` or they are absent, this is because they track the real state of the cluster resources monitored.

1. [ha_cluster_node_resources](#ha_cluster_node_resources)
2. [ha_cluster_nodes](#ha_cluster_nodes)
3. [ha_cluster_nodes_configured_total](#ha_cluster_nodes_configured_total)
4. [ha_cluster_resources_configured_total](#ha_cluster_resources_configured_total)



## ha_cluster_node_resources

This metric show the current status of a cluster resource. 

A resource that previously was in the cluster but isn't anymore, will not monitored. Example:

```ha_cluster_node_resources{managed="true",node_name="1b115",resource_name="cluster_md",role="started",status="active"} 1```

The metric will absent and not `0`


All the values are 1:1 with Pacemaker schema.

- `managed`: indicates `true` or `false` if the resource is managed in cluster
- `node_name`: name of node of cluster
- `resource_name`: resource id/name of the CIB pacemaker
- `role`:  allowed values `Started/Stopped/Master/Slave` or pending state `Starting/Stopping/Migrating/Promoting/Demoting` which are same as pacemaker roles for resources.
- `status` allowed values `active/orphaned/blocked/failed/failureIgnored/` status of resource from pacemaker XML.
           Additionaly for the same resource we can have a combination of status.

Example:

```
ha_cluster_node_resources{managed="true",node_name="1b115",resource_name="cluster_md",role="started",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b115",resource_name="clvm",role="started",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b115",resource_name="dlm",role="started",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b115",resource_name="drbd_passive",role="master",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b115",resource_name="fs_cluster_md",role="started",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b115",resource_name="fs_drbd_passive",role="started",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b115",resource_name="stonith-sbd",role="started",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b115",resource_name="vg_cluster_md",role="started",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b211",resource_name="dlm",role="started",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b211",resource_name="fs_cluster_md",role="stopped",status="active"} 1
ha_cluster_node_resources{managed="true",node_name="1b211",resource_name="vg_cluster_md",role="stopped",status="active"} 1
```

## ha_cluster_nodes

- `node_name`: name of cluster node
- `type`: allowed values  `online/standby/standby_onfail/maintanance/pending/unclean/shutdown/expected_up/dc/member/ping/remote/`. This are the possible type of pacemaker ha cluster

Again here, when the resource is absent will be not showed. There is no `0` value, since it is a real snapshot from the HA cluster.
Examples:
```
ha_cluster_nodes{node_name="1b115",type="dc"} 1
ha_cluster_nodes{node_name="1b115",type="expected_up"} 1
ha_cluster_nodes{node_name="1b115",type="member"} 1
ha_cluster_nodes{node_name="1b115",type="online"} 1
ha_cluster_nodes{node_name="1b211",type="expected_up"} 1
ha_cluster_nodes{node_name="1b211",type="member"} 1
ha_cluster_nodes{node_name="1b211",type="online"} 1
```

## ha_cluster_nodes_configured_total 

Show the total number of configured noded in the HA cluster

Example:

```
ha_cluster_nodes_configured_total 2
```


## ha_cluster_resources_configured_total 

Show the total number of resource configured in HA cluster
Example:
```
ha_cluster_resources_configured_total 14
```


# Corosync

`TODO`

# Drbd

`TODO`@MalloZup

# SBD

`TODO`
