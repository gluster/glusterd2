#!/bin/bash

GOPATH=$(go env GOPATH)
GOBINDIR=$GOPATH/bin

install_dep() {
  DEPVER="v0.5.0"
  DEPURL="https://github.com/golang/dep/releases/download/${DEPVER}/dep-linux-amd64"
  type dep >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    local version
    version=$(dep version | awk '/^ version/{print $3}')
    if [[ $version == "$DEPVER" || $version >  $DEPVER ]]; then
      echo "dep ${DEPVER} or greater is already installed"
      return
    fi
  fi

  echo "Installing dep. Version: ${DEPVER}"
  DEPBIN=$GOPATH/bin/dep
  curl -L -o "$DEPBIN" $DEPURL
  chmod +x "$DEPBIN"
}

install_gometalinter() {
  LINTER_VER="2.0.5"
  LINTER_TARBALL="gometalinter-${LINTER_VER}-linux-amd64.tar.gz"
  LINTER_URL="https://github.com/alecthomas/gometalinter/releases/download/v${LINTER_VER}/${LINTER_TARBALL}"

  type gometalinter >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "gometalinter already installed"
    return
  fi

  echo "Installing gometalinter. Version: ${LINTER_VER}"
  curl -L -o "$GOBINDIR/$LINTER_TARBALL" $LINTER_URL
  tar -zxf "$GOBINDIR/$LINTER_TARBALL" --overwrite --strip-components 1 --exclude={COPYING,*.md} -C "$GOBINDIR"
  rm -f "$GOBINDIR/$LINTER_TARBALL"
}

install_dep
install_gometalinter
