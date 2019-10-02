#/bin/bash

# THIS RUN ONLY WITH THE MAKEFILE TARGET make deploy

# this script is just for deploying the binary to the cluster. Nothing else.

node="root@10.162.31.221"

ssh $node "rm /root/ha_cluster_exporter"
echo "copying binary"
scp ha_cluster_exporter  $node:
