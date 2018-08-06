#!/bin/bash

REQ_GO_MAJOR_VERSION="1"
REQ_GO_MINOR_VERSION="9"

REQ_GO_VERSION="$REQ_GO_MAJOR_VERSION.$REQ_GO_MINOR_VERSION"

missing() {
  echo "Go$REQ_GO_VERSION or later is missing on this system."
  echo "Install Go${REQ_GO_VERSION} using the preferred method for your system."
  echo "Refer to https://golang.org/doc/install if Go$REQ_GO_VERSION is not available in the system repositories."

  exit 1
}

check_go_version() {

#Check if Go is installed
INST_VERS_STR=$(go version) || missing

INST_GO_VERSION=$(expr "$INST_VERS_STR" : ".*go version go\\([^ ]*\\) .*")
INST_GO_MAJOR_VERSION=$(echo "$INST_GO_VERSION" | cut -d. -f1)
INST_GO_MINOR_VERSION=$(echo "$INST_GO_VERSION" | cut -d. -f2)

if [ "$REQ_GO_MAJOR_VERSION" -gt "$INST_GO_MAJOR_VERSION" ]; then
        missing
elif [ "$REQ_GO_MAJOR_VERSION" -eq "$INST_GO_MAJOR_VERSION" ] &&
           [ "$REQ_GO_MINOR_VERSION" -gt "$INST_GO_MINOR_VERSION" ]; then
        missing
fi
}

check_go_version

echo "Go$REQ_GO_VERSION or later is available on the system."
