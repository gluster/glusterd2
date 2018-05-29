#!/bin/bash

# This scripts builds RPMs from the current git head.
# The script needs be run from the root of the repository
# NOTE: RPMs are built only for EL7 (CentOS7) distributions.

set -e

##
## Set up build environment
##
RESULTDIR=${RESULTDIR:-$PWD/rpms}
BUILDDIR=$PWD/$(mktemp -d nightlyrpmXXXXXX)

BASEDIR=$(dirname "$0")
GD2CLONE=$(realpath "$BASEDIR/..")

yum -y install make mock rpm-build golang

export GOPATH=$BUILDDIR/go
mkdir -p "$GOPATH"/{bin,pkg,src}
export PATH=$GOPATH/bin:$PATH

GD2SRC=$GOPATH/src/github.com/gluster/glusterd2
mkdir -p "$GOPATH/src/github.com/gluster"
ln -s "$GD2CLONE" "$GD2SRC"

"$GD2SRC/scripts/install-reqs.sh"

##
## Prepare GD2 archives and specfile for building RPMs
##
pushd "$GD2SRC"

VERSION=$(./scripts/pkg-version --version)
RELEASE=$(./scripts/pkg-version --release)
FULL_VERSION=$(./scripts/pkg-version --full)

# Create a vendored dist archive
DISTDIR=$BUILDDIR SIGN=no make dist-vendor

# Copy over specfile to the BUILDDIR and modify it to use the current Git HEAD versions
cp ./extras/rpms/* "$BUILDDIR"

popd #GD2SRC

pushd "$BUILDDIR"

DISTARCHIVE="glusterd2-$FULL_VERSION-vendor.tar.xz"
SPEC=glusterd2.spec
sed -i -E "
# Use bundled always
s/with_bundled 0/with_bundled 1/;
# Replace version with HEAD version
s/^Version:[[:space:]]+([0-9]+\\.)*[0-9]+$/Version: $VERSION/;
# Replace release with proper release
s/^Release:[[:space:]]+.*%\\{\\?dist\\}/Release: $RELEASE%{?dist}/;
# Replace Source0 with generated archive
s/^Source0:[[:space:]]+.*-vendor.tar.xz/Source0: $DISTARCHIVE/;
# Change prep setup line to use correct release
s/^(%setup -q -n %\\{name\\}-v%\\{version\\}-)(0)/\\1$RELEASE/;
" $SPEC

##
## Build the RPMs
##

# Create SRPM
mkdir -p rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
cp "$DISTARCHIVE" glusterd2-logrotate rpmbuild/SOURCES
cp $SPEC rpmbuild/SPECS
SRPM=$(rpmbuild --define "_topdir $PWD/rpmbuild" -bs rpmbuild/SPECS/$SPEC | cut -d\  -f2)

# Build RPM from SRPM using mock
mkdir -p "$RESULTDIR"
/usr/bin/mock -r epel-7-x86_64 --resultdir="$RESULTDIR" --rebuild "$SRPM"

popd #BUILDDIR

## Cleanup
rm -rf "$BUILDDIR"
