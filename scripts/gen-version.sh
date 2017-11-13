#!/bin/bash

# This scripts generates a VERSION file to be included in the dist tarball
# The generated VERSION file is imported and used by the build script

PKG_VERSION=$(git describe --tags --match "v[0-9]*")
GIT_SHA=$(git rev-parse --short HEAD)
GIT_SHA_FULL=$(git rev-parse HEAD)

cat >VERSION <<EOF
PKG_VERSION=$PKG_VERSION
GIT_SHA=$GIT_SHA
GIT_SHA_FULL=$GIT_SHA_FULL
EOF
