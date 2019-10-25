#
# spec file for package prometheus-ha_cluster_exporter
#
# Copyright (c) 2019 SUSE LINUX GmbH, Nuernberg, Germany.
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

# Please submit bugfixes or comments via http://bugs.opensuse.org/
#


Name:           prometheus-ha_cluster_exporter
Version:        0.0
Release:        0
License:        Apache-2.0
Summary:        Prometheus exporter for ha_cluster server metrics
Group:          System/Monitoring
Url:            https://github.com/ClusterLabs/ha_cluster_exporter
Source:         %{name}-%{version}.tar.gz
Source1:        vendor.tar.gz
BuildRequires:  git-core
BuildRequires:  fdupes
BuildRequires:  golang-packaging
BuildRequires:  go1.11
Provides:       ha_cluster_exporter = %{version}-%{release}
Provides:       prometheus(ha_cluster_exporter) = %{version}-%{release}
BuildRoot:      %{_tmppath}/%{name}-%{version}-build
%{go_provides}
# Make sure that the binary is not getting stripped.
%{go_nostrip}

%description
Prometheus exporter for ha_cluster pacemaker metrics.

%prep
%setup -q # unpack project sources
%setup -q -T -D -a 1 # unpack go dependencies in vendor.tar.gz, which was prepared by the source services

%define binary_name ha_cluster_exporter

%build
# we don't use OBS Go packaging macros but explicit go build command, as illustrated in the go_modules source service example
export VERSION=%{version}
export COMMIT=%{commit}
export CGO_ENABLED=0
go build \
   -mod=vendor \
   -buildmode=pie \
   -ldflags "-s -w -X main.gitCommit=$COMMIT -X main.version=$VERSION" \
   -o %{binary_name} ;

%install

# Install the binary.
install -D -m 0755 %{binary_name} "%{buildroot}%{_bindir}/%{binary_name}"

# Install the systemd unit
install -D -m 0644 %{name}.service %{buildroot}%{_unitdir}/%{name}.service

# Install compat wrapper for legacy init systems
install -Dd -m 0755 %{buildroot}%{_sbindir}
ln -s /usr/sbin/service %{buildroot}%{_sbindir}/rc%{name}

%fdupes %{buildroot}

%pre
%service_add_pre %{name}.service

%post
%service_add_post %{name}.service

%preun
%service_del_preun %{name}.service

%postun
%service_del_postun %{name}.service

%files
%defattr(-,root,root)
%doc *.md
%if 0%{?suse_version} >= 1500
%license LICENSE
%else
%doc LICENSE
%endif
%{_bindir}/%{binary_name}
%{_unitdir}/%{name}.service
%{_sbindir}/rc%{name}

%changelog
