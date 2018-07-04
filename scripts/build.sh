#!/usr/bin/env bash

## This scripts builds a GD2 binaries and places it in the given path
## Should be called from the root of the GD2 repo as
## ./scripts/build.sh <package> [<path-to-output-directory>]
## If no path is given, defaults to build

show_usage() {
  echo "Usage: $0 <package-path> [<output-directory>]"
  echo "<package-path>: Path of package to build relative to GD2 source root"
  echo "<output-directory>: Path of output directory. Defaults to 'build'"
  echo "Built binary will be placed at <output-directory>/<package-basename>)"
}


PACKAGE=${1}
if [[ "XX$PACKAGE" == "XX" ]]; then
  show_usage
  exit 1
fi

OUTDIR=${2:-build}
mkdir -p "$OUTDIR"

REPO_PATH="github.com/gluster/glusterd2"
GOPKG="${REPO_PATH}/${PACKAGE}"
BIN=$(basename "$PACKAGE")

VERSION=$("$(dirname "$0")/pkg-version" --full)
[[ -f VERSION ]] && source VERSION
GIT_SHA=${GIT_SHA:-$(git rev-parse --short HEAD || echo "undefined")}
GIT_SHA_FULL=${GIT_SHA_FULL:-$(git rev-parse HEAD || echo "undefined")}

LDFLAGS="-X ${REPO_PATH}/version.GlusterdVersion=${VERSION} -X ${REPO_PATH}/version.GitSHA=${GIT_SHA}"
LDFLAGS+=" -B 0x${GIT_SHA_FULL}"

if [ "$BIN" == "glusterd2" ]; then
    LDFLAGS+=" -X main.defaultPathPrefix=${BASE_PREFIX}"
fi

if [ "$BIN" == "glustercli" ]; then
    LDFLAGS+=" -X ${GOPKG}/cmd.defaultAuthPath=${GD2_STATE_DIR}/auth"
fi

GOBUILD_TAGS=""
if [ "$PLUGINS" == "yes" ]; then
    GOBUILD_TAGS+="plugins "
    echo "Plugins Enabled"
else
    echo "Plugins Disabled"
fi

if [ "$FASTBUILD" == "yes" ];then
  # Enable the `go build -i` flag to install dependencies during build and
  # allow faster rebuilds of GD2.
  INSTALLFLAG="-i"
fi


echo "Building $BIN $VERSION"

go build $INSTALLFLAG -ldflags "${LDFLAGS}" -o "$OUTDIR/$BIN" -tags "$GOBUILD_TAGS" "$GOPKG" || exit 1

echo "Built $PACKAGE $VERSION at $OUTDIR/$BIN"
