%global _dwz_low_mem_die_limit 0

%global provider github
%global provider_tld com
%global project gluster
%global repo glusterd2
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path %{provider_prefix}

Name: %{repo}
Version: 4.0dev
Release: 9
Summary: The GlusterFS management daemon (preview)
License: GPLv2 or LGPLv3+
URL: https://%{provider_prefix}
# Use vendored tarball instead of plain git archive
Source0: https://%{provider_prefix}/releases/download/v%{version}-%{release}/%{name}-v%{version}-%{release}-vendor.tar.gz
Source1: glusterd2.toml

ExclusiveArch: x86_64

BuildRequires: golang >= 1.8.0
BuildRequires: systemd

Requires: glusterfs-server >= 4.0dev
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

%description
Preview release of the next generation GlusterFS management framework and daemon, coming with GlusterFS-4.0

%prep
%setup -q -n %{name}-v%{version}-%{release}

%build
export GOPATH=$(pwd):%{gopath}

mkdir -p src/%(dirname %{import_path})
ln -s ../../../ src/%{import_path}

pushd src/%{import_path}
# Build glusterd2
make glusterd2
make glustercli
popd

%install
# TODO: Use make install to install
#Install glusterd2 & glustercli binary
install -D -p -m 0755 build/%{name} %{buildroot}%{_sbindir}/%{name}
install -D -p -m 0755 build/glustercli %{buildroot}%{_sbindir}/glustercli
#Install systemd unit
install -D -p -m 0644 extras/systemd/%{name}.service %{buildroot}%{_unitdir}/%{name}.service
#Install glusterd config into etc
install -d -m 0755 %{buildroot}%{_sysconfdir}/%{name}
install -m 0644 -t %{buildroot}%{_sysconfdir}/%{name} %{SOURCE1}
# Create /var/lib/glusterd2
install -d -m 0755 %{buildroot}%{_sharedstatedir}/%{name}
# logdir
install -d -m 0755 %{buildroot}%{_localstatedir}/log/%{name}
#Install templates
install -d -m 0755 %{buildroot}%{_datadir}/%{name}/templates
install -D -m 0644 -t %{buildroot}%{_datadir}/%{name}/templates volgen/templates/*.graph


%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun %{name}.service

%files
%{_sbindir}/%{name}
%{_sbindir}/glustercli
%config(noreplace) %{_sysconfdir}/%{glusterd2}
%{_unitdir}/%{name}.service
%dir %{_sharedstatedir}/%{name}
%dir %{_localstatedir}/log/%{name}
%dir %{_datadir}/%{name}
%dir %{_datadir}/%{name}/templates
%{_datadir}/%{name}/templates/*

%changelog
* Wed Nov 08 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-9
- Build with vendored tarball.

* Thu Oct 26 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-8
- Update spec file

* Mon Jul 03 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-7
- Initial spec
