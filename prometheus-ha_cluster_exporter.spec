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
Requires(post): %fillup_prereq
BuildRoot:      %{_tmppath}/%{name}-%{version}-build
%{go_provides}
# Make sure that the binary is not getting stripped.
%{go_nostrip}

#Compat macro for new _fillupdir macro introduced in Nov 2017
%if ! %{defined _fillupdir}
  %define _fillupdir /var/adm/fillup-templates
%endif
%description
Prometheus exporter for ha_cluster pacemaker metrics.
%prep
%setup -q -n %{name}-%{version}

%build
%goprep github.com/ClusterLabs/ha_cluster_exporter
%gobuild -mod=vendor ""

%install
%goinstall
%gosrc
install -D -m 0644 %{SOURCE1} %{buildroot}%{_unitdir}/%{name}.service
install -Dd -m 0755 %{buildroot}%{_sbindir}
ln -s /usr/sbin/service %{buildroot}%{_sbindir}/rc%{name}
%gofilelist
%fdupes %{buildroot}

%pre
%service_add_pre %{name}.service

%post
%service_add_post %{name}.service
%fillup_only -n %{name}

%preun
%service_del_preun %{name}.service

%postun
%service_del_postun %{name}.service

%files -f file.lst
%defattr(-,root,root)
%doc *.md
%if 0%{?suse_version} >= 1500
%license LICENSE
%else
%doc LICENSE
%endif
%{_bindir}/ha_cluster_exporter
%{_unitdir}/%{name}.service
%{_sbindir}/rc%{name}

%changelog
