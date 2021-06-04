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
Summary:        Grafana Dashboards for Pacemaker/Corosync HA Clusters
Group:          System/Monitoring
Url:            https://github.com/ClusterLabs/ha_cluster_exporter
Source:         %{name}-%{version}.tar.gz
BuildArch:      noarch
Requires(pre):  shadow
Requires:       grafana-sleha-provider
Recommends:     grafana

# TECHNICAL NOTE:
# Originally we used to require grafana but, for product management reasons, we use recommends now.
# This impacts how we do pkging here: requiring shadow, creating grafana usr/group
# and modifiying files attributes (this was done automagically when requiring grafana).

%description
Grafana Dashboards displaying metrics about Pacemaker/Corosync High Availability Clusters.

%package -n grafana-sleha-provider
Summary:        Grafana configuration providers for the SLES HA Extension
Group:          System/Monitoring
Recommends:     grafana
BuildArch:      noarch
Provides:       grafana-sleha-cluster-provider = %version-%release
Obsoletes:      grafana-sleha-cluster-provider < %version-%release

%description -n grafana-sleha-provider
Automated configuration provisioners leveraged by other packages to enable a zero-config installation of Grafana dashboards.

%prep
%setup -q

%pre
echo "Creating grafana user and group if not present"
getent group grafana > /dev/null || groupadd -r grafana
getent passwd grafana > /dev/null || useradd -r -g grafana -d  %{_datadir}/grafana -s /sbin/nologin grafana

%build

%install
install -d -m0755 %{buildroot}%{_localstatedir}/lib/grafana/dashboards/sleha
install -m644 dashboards/*.json %{buildroot}%{_localstatedir}/lib/grafana/dashboards/sleha
install -Dm644 dashboards/provider-sleha.yaml %{buildroot}%{_sysconfdir}/grafana/provisioning/dashboards/provider-sleha.yaml

%files
%defattr(-,root,root)
%doc dashboards/README.md
%license LICENSE
%attr(0644,grafana,grafana) %config %{_localstatedir}/lib/grafana/dashboards/sleha/*
%attr(0755,grafana,grafana) %dir %{_localstatedir}/lib/grafana
%attr(0755,grafana,grafana) %dir %{_localstatedir}/lib/grafana/dashboards
%attr(0755,grafana,grafana) %dir %{_localstatedir}/lib/grafana/dashboards/sleha

%files -n grafana-sleha-provider
%attr(0755,root,root) %dir %{_sysconfdir}/grafana
%attr(0755,root,root) %dir %{_sysconfdir}/grafana/provisioning
%attr(0755,root,root) %dir %{_sysconfdir}/grafana/provisioning/dashboards
%attr(0644,root,root) %config %{_sysconfdir}/grafana/provisioning/dashboards/provider-sleha.yaml

%changelog
