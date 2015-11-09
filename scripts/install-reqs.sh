#!/bin/bash

failed_install() {
  echo "Failed to install $1. Please install manually."
}

install_glide() {
  type glide >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "glide already installed"
    return
  fi

  echo "Installing glide"
  go get github.com/Masterminds/glide || failed_install glide
}

install_gometalinter() {
  type gometalinter >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "gometalinter already installed"
    return
  fi

  echo "Installing gometalinter"
  go get github.com/alecthomas/gometalinter
  if [ $? -ne 0 ]; then
    failed_install gometalinter
    return
  fi

  echo "Installing linters"
  gometalinter -i -u || failed_install linters
}

install_glide
install_gometalinter
