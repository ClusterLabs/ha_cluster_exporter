#/bin/bash

# this script is just for deploying the binary to the cluster. Nothing else.

node="root@10.162.31.230"

ssh $node "rm /root/ha_cluster_exporter"
echo "copying binary"
scp ../ha_cluster_exporter  root@10.162.31.230:

echo "run exporter"
ssh $node "/root/ha_cluster_exporter"
