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

%description
Grafana Dashboards displaying metrics about a Pacemaker/Corosync High Availability Cluster.

%prep
%setup -q

%build

%install
%define install_dir %{_sharedstatedir}/grafana/dashboards/sleha
install -d -m0755 %{buildroot}%{install_dir}
install -m644 dashboards/grafana-ha-cluster-details.json %{buildroot}%{install_dir}
install -m644 dashboards/grafana-multi-cluster-overview.json %{buildroot}%{install_dir}
install -m644 dashboards/provider-sleha.yaml %{_sysconfdir}/grafana/provisioning/dashboards

%files
%defattr(-,root,root)
%doc dashboard/README.md
%license LICENSE
%attr(0755,grafana,grafana) %dir %{install_dir}
%attr(0644,grafana,grafana) %config %{install_dir}/*
%attr(0755,root,root) %config %{_sysconfdir}/grafana/provisioning/dashboards/provider-sleha.yaml

%changelog
