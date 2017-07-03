#!/usr/bin/env bash

## This scripts builds a GD2 binary and places it in the given path
## Should be called from the root of the GD2 repo as
## ./scripts/build.sh [<path-to-output-directory>]
## If no path is given, defaults to build

OUTDIR=build
if [ "x$1" != "x" ]; then
  OUTDIR=$1
fi

GOBUILD_TAGS="novirt noaugeas "
VERSION=$($(dirname $0)/pkg-version --full)
LDFLAGS="-X github.com/gluster/glusterd2/gdctx.GlusterdVersion=$VERSION"
LDFLAGS+=" -B 0x$(head -c20 /dev/urandom | od -An -tx1 | tr -d ' \n')"
BIN=$(basename $(go list -f '{{.ImportPath}}'))

if [ "$PLUGINS" == "yes" ]; then
    GOBUILD_TAGS+="plugins "
    echo "Plugins Enabled"
else
    echo "Plugins Disabled"
fi

echo "Building $BIN $VERSION"

go build -ldflags "${LDFLAGS}" -o $OUTDIR/$BIN -tags "$GOBUILD_TAGS" || exit 1

echo "Built $BIN $VERSION at $OUTDIR/$BIN"
