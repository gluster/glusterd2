#!/usr/bin/env bash

## This script builds a GlusterD-2.0 binary and creates an archive, and then signs it.
## Should be called from the root of the GD2 repo

VERSION=$($(dirname $0)/pkg-version --full)
OS=$(go env GOOS)
ARCH=$(go env GOARCH)
BIN=$(basename $(go list -f '{{.ImportPath}}'))

RELEASEDIR=releases/$VERSION
TAR=$RELEASEDIR/$BIN-$VERSION-$OS-$ARCH.tar
ARCHIVE=$TAR.xz

if [ -e $ARCHIVE ]; then
  echo "Release archive $ARCHIVE exists."
  echo "Do you want to clean and start again?(y/N)"
  read answer
  case "$answer" in
    y|Y)
      echo "Cleaning previously built release"
      rm -rf $RELEASEDIR
      echo
      ;;
    *)
      exit 0
      ;;
  esac
fi

mkdir -p $RELEASEDIR

echo "Making GlusterD-2.0 release $VERSION"
echo

# Build GD2 into the release directory
$(dirname $0)/build.sh $RELEASEDIR || exit 1
echo

# Create release archive
echo "Creating release archive"
tar -cf $TAR -C $RELEASEDIR $BIN || exit 1
xz $TAR || exit 1
echo "Created release archive $RELEASEDIR/$ARCHIVE"
echo

# Sign the tarball
# Requires that a default gpg key be set up
echo "Signing archive"
SIGNFILE=$ARCHIVE.asc
gpg --armor --output $SIGNFILE --detach-sign $ARCHIVE || exit 1
echo "Signed archive, signature in $SIGNFILE"
