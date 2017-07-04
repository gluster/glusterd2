%global _dwz_low_mem_die_limit 0

%global provider github
%global provider_tld com
%global project gluster
%global repo glusterd2
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path %{provider_prefix}

Name: %{repo}
Version: 4.0dev
Release: 7
Summary: The GlusterFS management daemon (preview)
License: GPLv2 or LGPLv3+
URL: https://%{provider_prefix}
Source0: https://%{provider_prefix}/archive/v%{version}-%{release}/%{name}-v%{version}-%{release}.tar.gz
Source1: glusterd.toml

ExclusiveArch: x86_64

BuildRequires: golang >= 1.8.0
BuildRequires: glide >= 0.12.0
BuildRequires: git
BuildRequires: mercurial
BuildRequires: systemd

Requires: glusterfs-server >= 3.11.0
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

%description
Preview release of the next generation GlusterFS management framework and daemon, coming with GlusterFS-4.0

%prep
%setup -q -n %{name}-v%{version}-%{release}

%build
mkdir -p src/%(dirname %{import_path})
ln -s ../../../ src/%{import_path}

pushd src/%{import_path}
# Install vendored packages
# TODO: See if we can build with unbundled packages
make vendor-install
# Build glusterd2
make glusterd2
popd

%install
#Install glusterd2 binary
install -D -p -m 0755 build/%{name} %{buildroot}%{_sbindir}/%{name}
#Install systemd unit
install -D -p -m 0644 extras/systemd/%{name}.service %{buildroot}%{_unitdir}/%{name}.service
#Install glusterd config into etc
install -d -m 0755 %{buildroot}%{_sysconfdir}/glusterd
install -m 0644 -t %{buildroot}%{_sysconfdir}/glusterd %{SOURCE1}
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
%config(noreplace) %{_sysconfdir}/glusterd
%{_unitdir}/%{name}.service
%dir %{_sharedstatedir}/%{name}
%dir %{_localstatedir}/log/%{name}

%changelog
* Mon Jul 03 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-7
- Initial spec
