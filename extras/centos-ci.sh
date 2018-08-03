#!/bin/bash

# This script will be called by the gluster_glusterd2 job script on centos-ci.
# This script sets up the centos-ci environment and runs the PR tests for GD2.

# if anything fails, we'll abort
set -e

REQ_GO_VERSION='1.9.4'
# install Go
if ! yum -y install "golang >= $REQ_GO_VERSION"
then
	# not the right version, install manually
	# download URL comes from https://golang.org/dl/
	curl -O https://storage.googleapis.com/golang/go${REQ_GO_VERSION}.linux-amd64.tar.gz
	tar xzf go${REQ_GO_VERSION}.linux-amd64.tar.gz -C /usr/local
	export PATH=$PATH:/usr/local/go/bin
fi

# also needs git, hg, bzr, svn gcc and make
yum -y install git mercurial bzr subversion gcc make

# Install nightly GlusterFS RPMs built off master
curl -o /etc/yum.repos.d/glusterfs-nighthly-master.repo http://artifacts.ci.centos.org/gluster/nightly/master.repo
yum -y install epel-release
yum -y install glusterfs-server
yum -y install ShellCheck

export GD2SRC=$GOPATH/src/github.com/gluster/glusterd2
cd "$GD2SRC"

# install the build and test requirements
./scripts/install-reqs.sh

# install vendored dependencies
make vendor-install

# verify build
make glusterd2
make glustercli
make gd2conf

# run tests
make test TESTOPTIONS=-v

# run functional tests
make functest
