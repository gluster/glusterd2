#!/usr/bin/env bash

# This script builds a dist tarball of the GD2 source
# This should only be called from the root of the GD2 repo

VENDOR=${VENDOR:-no}
OUTDIR=${1:-.}

VERSION=$($(dirname $0)/pkg-version --full)

TARNAME=glusterd2-$VERSION
case $VENDOR in
  yes|y|Y)
    TARNAME+="-vendor"
    ;;

  *)
    EXCLUDEVENDOR="--exclude vendor"
    ;;
esac

TARBALL=$TARNAME.tar
ARCHIVE=$OUTDIR/$TARBALL.xz
SIGNFILE=$ARCHIVE.asc

# Cleanup old archives
if [[ -f $ARCHIVE ]]; then
  rm $ARCHIVE
fi
if [[ -f $SIGNFILE ]]; then
  rm $SIGNFILE
fi

# Create the VERSION file first
$(dirname $0)/gen-version.sh

# Use a temp dir to build tar file
TMPDIR=$(mktemp -d)

echo "Creating dist tarball $ARCHIVE"
# Exclude .git, releases, builds and older tarballs
tar --exclude .git --exclude releases --exclude build --exclude 'glusterd2-*.tar.xz*' $EXCLUDEVENDOR --transform "s/^\./$TARNAME/" -cf $TMPDIR/$TARBALL . || exit 1
xz --stdout $TMPDIR/$TARBALL > $ARCHIVE || exit 1
echo "Created dist tarball $ARCHIVE"


# Sign the generated archive
echo "Signing dist tarball"
gpg --armor --output $SIGNFILE --detach-sign $ARCHIVE || exit 1
echo "Signed dist tarball, signature in $SIGNFILE"


# Remove the VERSION file, it is no longer needed, and TMPDIR
rm VERSION
rm -rf $TMPDIR

