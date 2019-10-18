#!/usr/bin/env bash

cat <<EOF
Printing ring status.
Local node ID 16777226
RING ID 0
    id      = 10.0.0.1
    status  = Marking ringid 0 interface 10.0.0.1 FAULTY
RING ID 1
    id      = 172.16.0.1
    status  = ring 1 active with no faults
EOF
