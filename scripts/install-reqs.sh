#!/bin/bash

GOPATH=$(go env GOPATH)

failed_install() {
  echo "Failed to install $1. Please install manually."
}

install_dep() {
  DEPVER="v0.3.1"
  DEPURL="https://github.com/golang/dep/releases/download/${DEPVER}/dep-linux-amd64"
  type dep >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    local version=$(dep version | awk '/^ version/{print $3}')
    if [[ $version == $DEPVER || $version >  $DEPVER ]]; then
      echo "dep ${DEPVER} or greater is already installed"
      return
    fi
  fi

  echo "Installing dep"
  DEPBIN=$GOPATH/bin/dep
  curl -L -o $DEPBIN $DEPURL
  chmod +x $DEPBIN
}

install_gometalinter() {
  type gometalinter >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "gometalinter already installed"
    return
  fi

  echo "Installing gometalinter"
  go get -u github.com/alecthomas/gometalinter
  if [ $? -ne 0 ]; then
    failed_install gometalinter
    return
  fi

  echo "Installing linters"
  gometalinter --install --update || failed_install linters
}

install_dep
install_gometalinter
