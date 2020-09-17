#
# Copyright 2019-2020 SUSE LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
Name:           grafana-ha-cluster-dashboards
# Version will be processed via set_version source service
Version:        0
Release:        0
License:        Apache-2.0
Summary:        Grafana Dashboards displaying metrics about a Pacemaker/Corosync High Availability Cluster.
Group:          System/Monitoring
Url:            https://github.com/ClusterLabs/ha_cluster_exporter
Source:         %{name}-%{version}.tar.gz
BuildArch:      noarch
Requires(pre):  shadow
Recommends:     grafana

# TECHNICAL NOTE:
# Originally we were requiring grafana pkg. For Distros reasons, we use recommends
# this impact how we do pkging here: requireing shadow, creating grafana usr/group
# and modifiying owning the directories. ( this was done automagically when requiring grafana)
%description
Grafana Dashboards displaying metrics about a Pacemaker/Corosync High Availability Cluster.

%prep
%setup -q

%pre
echo "Creating grafana user and group if not present"
getent group grafana > /dev/null || groupadd -r grafana
getent passwd grafana > /dev/null || useradd -r -g grafana -d  %{_datadir}/grafana -s /sbin/nologin grafana

%build

%install
%define dashboards_dir %{_localstatedir}/lib/grafana/dashboards
%define provisioning_dir %{_sysconfdir}/grafana/provisioning/dashboards
install -d -m0755 %{buildroot}%{dashboards_dir}/sleha
install -m644 dashboards/*.json %{buildroot}%{dashboards_dir}/sleha
install -Dm644 dashboards/provider-sleha.yaml %{buildroot}%{provisioning_dir}/provider-sleha.yaml

%files
%defattr(-,root,root)
%doc dashboards/README.md
%license LICENSE
%attr(0755,grafana,grafana) %dir %{dashboards_dir}/sleha
%attr(0644,grafana,grafana) %config %{dashboards_dir}/sleha/*
%attr(0644,grafana,grafana) %config %{provisioning_dir}/provider-sleha.yaml
%attr(0755,root,root) %dir  %{_sysconfdir}/grafana
%attr(0755,root,root) %dir  %{_sysconfdir}/grafana/provisioning
%attr(0755,root,root) %dir  %{_sysconfdir}/grafana/provisioning/dashboards
%attr(0755,grafana,grafana) %dir  %{_localstatedir}/lib/grafana
%attr(0755,grafana,grafana) %dir  %{_localstatedir}/lib/grafana/dashboards

%changelog
