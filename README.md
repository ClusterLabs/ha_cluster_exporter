# ha_cluster_exporter

[![Exporter CI](https://github.com/ClusterLabs/ha_cluster_exporter/workflows/Exporter%20CI/badge.svg)](https://github.com/ClusterLabs/ha_cluster_exporter/actions?query=workflow%3A%22Exporter+CI%22)
[![Dashboards CI](https://github.com/ClusterLabs/ha_cluster_exporter/workflows/Dashboards%20CI/badge.svg)](https://github.com/ClusterLabs/ha_cluster_exporter/actions?query=workflow%3A%22Dashboards+CI%22)

This is a bespoke Prometheus exporter used to enable the monitoring of Pacemaker based HA clusters.  

## Table of Contents
1. [Features](#features)
2. [Installation](#installation)
3. [Usage](#usage)
   1. [Metrics](doc/metrics.md)
   2. [Dashboards](dashboards/README.md)
5. [Contributing](#contributing)
   1. [Design](doc/design.md)
   2. [Development](doc/development.md)
5. [License](#license)

## Features

The exporter is a stateless HTTP endpoint. On each HTTP request, it locally inspects the cluster status by parsing pre-existing distributed data, provided by the tools of the various cluster components.

Exported data include:
- Pacemaker cluster summary, nodes and resources stats 
- Corosync ring errors and quorum votes
- SBD devices health status 
- DRBD resources and connections stats  
  (note: only DBRD v9 is supported; for v8.4, please refer to the [Prometheus Node Exporter](https://github.com/prometheus/node_exporter) project)

A comprehensive list of all the metrics can be found in the [metrics document](doc/metrics.md).

## Installation

The project can be installed in many ways, including but not limited to:

1. [Manual clone & build](#manual-clone-&-build)
2. [Go](#go)
3. [RPM](#rpm)
4. [Docker](#docker)

### Manual clone & build

```
git clone https://github.com/ClusterLabs/ha_cluster_exporter
cd ha_cluster_exporter
make
make install
```

### Go

```
go install github.com/ClusterLabs/ha_cluster_exporter/cmd/ha_cluster_exporter@latest
```

### RPM

On openSUSE or SUSE Linux Enterprise you can just use the `zypper` system package manager:
```shell
zypper install prometheus-ha_cluster_exporter
```

You can find the latest development repositories at [SUSE's Open Build Service](https://build.opensuse.org/package/show/network:ha-clustering:sap-deployments:devel/prometheus-ha_cluster_exporter).

### Docker

You can build and run the exporter as a Docker container:

```shell
make docker
docker run -p 9664:9664 ha_cluster_exporter
```

Note: To collect metrics from the host cluster, the container may need access to the host's network and tools.

## Usage

You can run the exporter in any of the cluster nodes. 

```
$ ./ha_cluster_exporter  
INFO[0000] Serving metrics on 0.0.0.0:9664
```

Though not strictly required, it is _strongly_ advised to run it in all the nodes.

It will export the metrics under the `/metrics` path, on port `9664` by default.

While the exporter can run outside a HA cluster node, it won't export any metric it can't collect; e.g. it won't export DRBD metrics if it can't be locally inspected with `drbdsetup`.  
A warning message will inform the user of such cases.

Please, refer to [doc/metrics.md](doc/metrics.md) for extensive details about all the exported metrics.

To see a practical example of how to consume the metrics, we also provide a couple of [Grafana dashboards](dashboards). 

**Hint:**
You can deploy a full HA Cluster via Terraform with [SUSE/ha-sap-terraform-deployments](https://github.com/SUSE/ha-sap-terraform-deployments).

### Configuration

All the runtime parameters can be configured either via CLI flags or via a configuration file, both or which are completely optional.

For more details, refer to the help message via `ha_cluster_exporter --help`.

**Note**:
the built-in defaults are tailored for the latest version of SUSE Linux Enterprise and openSUSE.

The program will scan, in order, the current working directory, `$HOME/.config`, `/etc` and `/usr/etc` for files named `ha_cluster_exporter.(yaml|json|toml)`.
The first match has precedence, and the CLI flags have precedence over the config file.

Please refer to the example [YAML configuration](ha_cluster_exporter.yaml) for more details.

Additional CLI flags can also be passed via `/etc/sysconfig/prometheus-ha_cluster_exporter`.

#### General Flags

Name                                       | Description
----                                       | -----------
web.listen-address                         | Address to listen on for web interface and telemetry (default `:9664`).
web.telemetry-path                         | Path under which to expose metrics (default `/metrics`).
web.config.file                            | Path to a [web configuration file](#tls-and-basic-authentication) (default `/etc/ha_cluster_exporter.web.yaml`).
log.level                                  | Logging verbosity (default `info`).
version                                    | Print the version information.

#### Collector Flags

Name                                       | Description
----                                       | -----------
collector.pacemaker                        | Enable the Pacemaker collector (default: enabled).
collector.corosync                         | Enable the Corosync collector (default: enabled).
collector.sbd                              | Enable the SBD collector (default: enabled).
collector.drbd                             | Enable the DRBD collector (default: enabled).
collector.timeout                          | Timeout for system commands execution (default `10s`).
crm-mon-path                               | Path to crm_mon executable (default `/usr/sbin/crm_mon`).
cibadmin-path                              | Path to cibadmin executable (default `/usr/sbin/cibadmin`).
corosync-cfgtoolpath-path                  | Path to corosync-cfgtool executable (default `/usr/sbin/corosync-cfgtool`).
corosync-quorumtool-path                   | Path to corosync-quorumtool executable (default `/usr/sbin/corosync-quorumtool`).
sbd-path                                   | Path to sbd executable (default `/usr/sbin/sbd`).
sbd-config-path                            | Path to sbd configuration (default `/etc/sysconfig/sbd`).
drbdsetup-path                             | Path to drbdsetup executable (default `/sbin/drbdsetup`).
drbdsplitbrain-path                        | Path to drbd splitbrain hooks temporary files (default `/var/run/drbd/splitbrain`).

### TLS and basic authentication

The ha_cluster_exporter supports TLS and basic authentication.

To use TLS and/or basic authentication, you need to pass a configuration file
using the `--web.config.file` parameter. The format of the file is described
[in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

### systemd integration

A [systemd unit file](ha_cluster_exporter.service) is provided with the RPM packages. You can enable and start it as usual:  

```
systemctl --now enable prometheus-ha_cluster_exporter
```

## Development

Pull requests are more than welcome!

We recommend having a look at the [design document](doc/design.md) and the [development notes](doc/development.md) before contributing.

## License

Copyright 2019-2022 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
