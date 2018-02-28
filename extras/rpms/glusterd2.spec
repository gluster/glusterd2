%if 0%{fedora}
# Feb 14 2018 - GD2 does not build with the already available Go package
# dependencies in Fedora. So until they are updated, build with the bundled
# dependencies
# TODO: Disable bundled builds once dependencies are met
%global with_bundled 1
%else
%global with_bundled 1
%endif

%if 0%{?with_debug}
%global _dwz_low_mem_die_limit 0
%else
%global debug_package   %{nil}
%endif

%global provider github
%global provider_tld com
%global project gluster
%global repo glusterd2
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path %{provider_prefix}

%global gd2make %{__make} PREFIX=%{_prefix} EXEC_PREFIX=%{_exec_prefix} BINDIR=%{_bindir} SBINDIR=%{_sbindir} DATADIR=%{_datadir} LOCALSTATEDIR=%{_sharedstatedir} LOGDIR=%{_localstatedir}/log SYSCONFDIR=%{_sysconfdir} FASTBUILD=off

Name: %{repo}
Version: 4.0.0
Release: 1%{?dist}
Summary: The GlusterFS management daemon (preview)
License: GPLv2 or LGPLv3+
URL: https://%{provider_prefix}
%if 0%{?with_bundled}
Source0: https://%{provider_prefix}/releases/download/v%{version}/%{name}-v%{version}-0-vendor.tar.xz
%else
Source0: https://%{provider_prefix}/releases/download/v%{version}/%{name}-v%{version}-0.tar.xz
%endif

ExclusiveArch: %{go_arches}

BuildRequires: golang >= 1.8.0
BuildRequires: systemd
%if ! 0%{?with_bundled}
BuildRequires: golang(github.com/asaskevich/govalidator)
BuildRequires: golang(github.com/cespare/xxhash)
BuildRequires: golang(github.com/cockroachdb/cmux)
BuildRequires: golang(github.com/coreos/etcd/clientv3)
BuildRequires: golang(github.com/coreos/etcd/clientv3/concurrency)
BuildRequires: golang(github.com/coreos/etcd/clientv3/namespace)
BuildRequires: golang(github.com/coreos/etcd/embed)
BuildRequires: golang(github.com/coreos/etcd/etcdserver/etcdserverpb)
BuildRequires: golang(github.com/coreos/etcd/pkg/transport)
BuildRequires: golang(github.com/coreos/etcd/pkg/types)
BuildRequires: golang(github.com/coreos/pkg/capnslog)
BuildRequires: golang(github.com/dgrijalva/jwt-go)
BuildRequires: golang(github.com/golang/protobuf/proto)
BuildRequires: golang(github.com/gorilla/handlers)
BuildRequires: golang(github.com/gorilla/mux)
BuildRequires: golang(github.com/justinas/alice)
BuildRequires: golang(github.com/olekukonko/tablewriter)
BuildRequires: golang(github.com/pborman/uuid)
BuildRequires: golang(github.com/pelletier/go-toml)
BuildRequires: golang(github.com/rasky/go-xdr/xdr2)
BuildRequires: golang(github.com/sirupsen/logrus)
BuildRequires: golang(github.com/spf13/cobra)
BuildRequires: golang(github.com/spf13/pflag)
BuildRequires: golang(github.com/spf13/viper)
BuildRequires: golang(github.com/thejerf/suture)
BuildRequires: golang(golang.org/x/net/context)
BuildRequires: golang(golang.org/x/sys/unix)
BuildRequires: golang(google.golang.org/grpc)
%endif

Requires: glusterfs-server >= 4.0.0
Requires: /usr/bin/strings
%{?systemd_requires}

%description
The new GlusterFS management framework and daemon, for GlusterFS-4.0.

%prep
%setup -q -n %{name}-v%{version}-0

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
* Wed Feb 28 2018 Kaushal M <kshlmster@gmail.com> - 4.0.0
- Update to v4.0.0

* Wed Feb 14 2018 Kaushal M <kshlmster@gmail.com> - 4.0rc0-2
- Update spec to support unbundled/vendored builds
- Fedora defaults to bundled builds till all required dependencies are available

* Tue Jan 30 2018 Kaushal M <kshlmster@gmail.com> - 4.0rc0-1
- Switch ExclusiveArch to go_arches

* Fri Jan 12 2018 Kaushal M <kshlmster@gmail.com> - 4.0dev-10
- Use standard paths to build and install

* Wed Nov 08 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-9
- Build with vendored tarball.

* Thu Oct 26 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-8
- Update spec file

* Mon Jul 03 2017 Kaushal M <kshlmster@gmail.com> - 4.0dev-7
- Initial spec
