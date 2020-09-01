#!/usr/bin/env bash

cat <<EOF
==Dumping header on disk /dev/vdc
Header version     : 2.1
UUID               : 1ed3171d-066d-47ca-8f76-aec25d9efed4
Number of slots    : 255
Sector size        : 512
Timeout (watchdog) : 9
Timeout (allocate) : 2
Timeout (loop)     : 1
Timeout (msgwait)  : 10
==Header on disk /dev/vdc is dumped
EOF
