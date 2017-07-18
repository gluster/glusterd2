#!/bin/bash

# This script will be called by the gluster_glusterd2 job script on centos-ci.
# This script sets up the centos-ci environment and runs the PR tests for GD2.

# if anything fails, we'll abort
set -e

REQ_GO_VERSION='1.8.0'
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

export GD2SRC=$GOPATH/src/github.com/gluster/glusterd2
cd $GD2SRC

# install the build and test requirements
./scripts/install-reqs.sh

# install vendored dependencies
make vendor-install

# run linters
make verify

# verify build
make glusterd2

# run unit-tests
make test
