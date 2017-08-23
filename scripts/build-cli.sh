#!/usr/bin/env bash

## This scripts builds a GD2 CLI binary and places it in the given path
## Should be called from the root of the GD2 repo as
## ./scripts/build-cli.sh [<path-to-output-directory>]
## If no path is given, defaults to build dir

OUTDIR=build
if [ "x$1" != "x" ]; then
  OUTDIR=$1
fi

GOBUILD_TAGS="novirt noaugeas "
VERSION=$($(dirname $0)/pkg-version --full)
REPO_PATH="github.com/gluster/glusterd2"
GIT_SHA=`git rev-parse --short HEAD || echo "undefined"`
LDFLAGS="-X ${REPO_PATH}/version.GlusterdVersion=$VERSION -X ${REPO_PATH}/version.GitSHA=$GIT_SHA"
LDFLAGS+=" -B 0x$(head -c20 /dev/urandom | od -An -tx1 | tr -d ' \n')"
BIN=glustercli

echo "Building $BIN $VERSION"

cd cli
go build -ldflags "${LDFLAGS}" -o ../$OUTDIR/$BIN -tags "$GOBUILD_TAGS" || exit 1
cd ..

echo "Built $BIN $VERSION at $OUTDIR/$BIN"
