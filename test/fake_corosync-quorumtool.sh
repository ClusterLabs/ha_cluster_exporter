#!/usr/bin/env bash

cat <<EOF
Quorum information
------------------
Date:             Fri Oct 18 12:46:58 2019
Quorum provider:  corosync_votequorum
Nodes:            2
Node ID:          1084783375
Ring ID:          1084783375/40
Quorate:          Yes

Votequorum information
----------------------
Expected votes:   2
Highest expected: 2
Total votes:      2
Quorum:           1
Flags:            2Node Quorate

Membership information
----------------------
    Nodeid      Votes  Qdevice Name
1084783375          1      NR  stefanotorresi-hana01 (local)
1084783376          1  A,V,NMW stefanotorresi-hana02
         0          1            Qdevice
EOF
