#!/usr/bin/env bash

## This script builds a GlusterD-2.0 binary and creates an archive, and then signs it.
## Should be called from the root of the GD2 repo

VERSION=$("$(dirname "$0")"/pkg-version --full)
OS=$(go env GOOS)
ARCH=$(go env GOARCH)
GD2=glusterd2

RELEASEDIR=releases/$VERSION
TAR=$RELEASEDIR/$GD2-$VERSION-$OS-$ARCH.tar
ARCHIVE=$TAR.xz

TMPDIR=$(mktemp -d)

if [ -e "$ARCHIVE" ]; then
  echo "Release archive $ARCHIVE exists."
  echo "Do you want to clean and start again?(y/N)"
  read -r answer
  case "$answer" in
    y|Y)
      echo "Cleaning previously built release"
      rm -rf "$RELEASEDIR"
      echo
      ;;
    *)
      exit 0
      ;;
  esac
fi

mkdir -p "$RELEASEDIR"

echo "Making GlusterD-2.0 release $VERSION"
echo

cp build/glusterd2 "$TMPDIR"
cp build/glustercli "$TMPDIR"
cp build/glusterd2.toml "$TMPDIR/glusterd2.toml.example"
mkdir "$TMPDIR/bash_completion"
cp build/glustercli.sh "$TMPDIR/bash_completion/glustercli.sh"
echo

# Create release archive
echo "Creating release archive"
tar -cf "$TAR" -C "$TMPDIR" . || exit 1
xz "$TAR" || exit 1
echo "Created release archive $RELEASEDIR/$ARCHIVE"
echo

# Sign the tarball
# Requires that a default gpg key be set up
echo "Signing archive"
SIGNFILE=$ARCHIVE.asc
gpg --armor --output "$SIGNFILE" --detach-sign "$ARCHIVE" || exit 1
echo "Signed archive, signature in $SIGNFILE"

rm -rf "$TMPDIR"

# Also create source tarballs
DISTDIR="$RELEASEDIR" "$(dirname "$0")/dist.sh"
VENDOR=y DISTDIR="$RELEASEDIR" "$(dirname "$0")/dist.sh"
