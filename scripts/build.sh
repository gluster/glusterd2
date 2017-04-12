#!/usr/bin/env bash

## This scripts builds a GD2 binary and places it in the given path
## Should be called from the root of the GD2 repo as
## ./scripts/build.sh [<path-to-output-directory>]
## If no path is given, defaults to build

OUTDIR=build
if [ "x$1" != "x" ]; then
  OUTDIR=$1
fi

VERSION=$($(dirname $0)/pkg-version --full)
LDFLAGS="-X github.com/gluster/glusterd2/gdctx.GlusterdVersion=$VERSION"
BIN=$(basename $(go list -f '{{.ImportPath}}'))

echo "Building $BIN $VERSION"
go build -ldflags "${LDFLAGS}" -tags "noaugeas novirt" -o $OUTDIR/$BIN || exit 1
echo "Built $BIN $VERSION at $OUTDIR/$BIN"


