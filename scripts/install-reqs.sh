#!/bin/bash

failed_install() {
  echo "Failed to install $1. Please install manually."
}

install_glide() {
  GLIDEVER="v0.12.3"
  GLIDEURL="https://github.com/Masterminds/glide/releases/download/${GLIDEVER}/glide-${GLIDEVER}-linux-amd64.tar.gz"
  type glide >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    local version=$(glide --version | awk '{print $3}')
    if [[ $version == $GLIDEVER || $version >  $GLIDEVER ]]; then
      echo "glide $GLIDEVER or greater is already installed"
      return
    fi
  fi

  echo "Installing glide"
  TMPD=$(mktemp -d)
  pushd $TMPD
  curl -L -o glide.tar.gz $GLIDEURL
  tar zxf glide.tar.gz
  cp linux-amd64/glide $GOPATH/bin
  popd
  rm -rf $TMPD
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

install_glide
install_gometalinter
