#!/usr/bin/env bash

# This script builds a dist tarball of the GD2 source
# This should only be called from the root of the GD2 repo

VENDOR=${VENDOR:-no}
OUTDIR=${DISTDIR:-.}
SIGN=${SIGN:-yes}

VERSION=$("$(dirname "$0")/pkg-version" --full)

BASENAME=glusterd2-$VERSION
TARNAME=$BASENAME
case $VENDOR in
  yes|y|Y)
    TARNAME+="-vendor"
    ;;
esac

TARFILE=$OUTDIR/$TARNAME.tar
ARCHIVE=$TARFILE.xz
SIGNFILE=$ARCHIVE.asc

# Cleanup old archives
if [[ -f $ARCHIVE ]]; then
  rm "$ARCHIVE"
fi
if [[ -f $SIGNFILE ]]; then
  rm "$SIGNFILE"
fi

# Create the VERSION file first
"$(dirname "$0")/gen-version.sh"

echo "Creating dist archive $ARCHIVE"
git archive -o "$TARFILE" --prefix "$BASENAME/" HEAD
tar --transform "s/^\\./$BASENAME/" -rf "$TARFILE" ./VERSION || exit 1
case $VENDOR in
  yes|y|Y)
    tar --transform "s/^\\./$BASENAME/" -rf "$TARFILE" ./vendor || exit 1
    ;;
esac

xz "$TARFILE" || exit 1
echo "Created dist archive $ARCHIVE"


# Sign the generated archive
case $SIGN in
  yes|y|Y)
    echo "Signing dist archive"
    gpg --armor --output "$SIGNFILE" --detach-sign "$ARCHIVE" || exit 1
    echo "Signed dist archive, signature in $SIGNFILE"
    ;;
esac

# Remove the VERSION file, it is no longer needed and would harm normal builds
rm VERSION

