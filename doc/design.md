# Design:

The design of the exporter is based on commit : https://github.com/MalloZup/ha_cluster_exporter/commit/7961a8a8147987977af1a32d8ec1ec01826b8b83
Some principle might have changed but not much.

![design](design.jpeg)

First in the `main` function we setup the Prometheus exporter metrics constructs. A metric hold a state. You can imagine them as global mutable variables which are served over http at the end.

The main functionality of the exporter is executed in a golang routine. At the begin of the loop most of all metrics are "reset", so all the old information/state is removed. 
(this is done to clean up metrics who have complex labels)

The `cluster` state is retrieved by the  `crm_mon` command. Since the data is XML struct, the `crm_mon` is just a connector. Other commands could be used to get the state as "connector".

Once the data is retrieved, this the golang types are popoulated. 

Within this we set the various metrics (gauge, gaugeVec). 

THe loops sleep a X timeout . And go to being

The metric are served via http
