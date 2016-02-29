#!/bin/bash

failed_install() {
  echo "Failed to install $1. Please install manually."
}

install_glide() {
  GLIDEVER="0.9.1"
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
  wget --quiet -O glide.tar.gz $GLIDEURL
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

install_etcd() {
        ETCDVERSION="v2.2.4"
        ETCDURL="https://github.com/coreos/etcd/releases/download/${ETCDVERSION}/etcd-${ETCDVERSION}-linux-amd64.tar.gz"
        type etcd >/dev/null 2>&1
        if [ $? -eq 0 ]; then
                echo "etcd already installed"
                return
        fi

        echo "Installing ETCD version ${ETCDVERSION}"
        TMPD=$(mktemp -d)
        pushd $TMPD
        echo ${TMPD}
        curl -L $ETCDURL -o etcd-${VERSION}-linux-amd64.tar.gz

        tar xzvf etcd-${VERSION}-linux-amd64.tar.gz

        cp etcd-${ETCDVERSION}-linux-amd64/etcd     $GOPATH/bin
        cp etcd-${ETCDVERSION}-linux-amd64/etcdctl  $GOPATH/bin
        popd
        rm -rf $TMPD
}

install_glide
install_gometalinter
install_etcd
