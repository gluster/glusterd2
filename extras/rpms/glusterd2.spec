%global _dwz_low_mem_die_limit 0

%global provider github
%global provider_tld com
%global project gluster
%global repo glusterd2
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path %{provider_prefix}

%global gd2make %{__make} PREFIX=%{_prefix} EXEC_PREFIX=%{_exec_prefix} BINDIR=%{_bindir} SBINDIR=%{_sbindir} DATADIR=%{_datadir} LOCALSTATEDIR=%{_sharedstatedir} LOGDIR=%{_localstatedir}/log SYSCONFDIR=%{_sysconfdir}

Name: %{repo}
Version: 4.0rc0
Release: 1
Summary: The GlusterFS management daemon (preview)
License: GPLv2 or LGPLv3+
URL: https://%{provider_prefix}
# Use vendored tarball instead of plain git archive
Source0: https://%{provider_prefix}/releases/download/v%{version}-%{release}/%{name}-v%{version}-%{release}-vendor.tar.xz

ExclusiveArch: %{go_arches}

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
%{gd2make} glusterd2
%{gd2make} glustercli
%{gd2make} glusterd2.toml
popd

%install
#Install glusterd2 & glustercli binaries and the config
%{gd2make} DESTDIR=%{buildroot} install
#Install systemd unit
install -D -p -m 0644 extras/systemd/%{name}.service %{buildroot}%{_unitdir}/%{name}.service
# Create /var/lib/glusterd2
install -d -m 0755 %{buildroot}%{_sharedstatedir}/%{name}
# logdir
install -d -m 0755 %{buildroot}%{_localstatedir}/log/%{name}

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun %{name}.service

%files
%{_sbindir}/%{name}
%{_sbindir}/glustercli
%config(noreplace) %{_sysconfdir}/%{name}
%{_unitdir}/%{name}.service
%dir %{_sharedstatedir}/%{name}
%dir %{_localstatedir}/log/%{name}
%{_sysconfdir}/bash_completion.d/glustercli.sh

%changelog
* Tue Jan 30 2018 Kaushal M <kshlmster@gmail.com> - 4.0rc0
- Switch ExclusiveArch to go_arches

* Fri Jan 12 2018 Kaushal M <kshlmster@gmail.com> - 4.0dev-10
- Use standard paths to build and install

* Wed Nov 08 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-9
- Build with vendored tarball.

* Thu Oct 26 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-8
- Update spec file

* Mon Jul 03 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-7
- Initial spec
