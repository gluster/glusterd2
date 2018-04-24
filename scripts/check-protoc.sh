#!/bin/bash

# Protobuf 3 version of `protoc` is required for GD2,
# for generating Go code from the proto definitions.

if [ "x$PROTOC" == "x" ]; then
  PROTOC=protoc
fi

REQ_PROTOC_MAJOR_VERSION="3"
REQ_PROTOC_MINOR_VERSION="0"

REQ_PROTOC_VERSION="$REQ_PROTOC_MAJOR_VERSION.$REQ_PROTOC_MINOR_VERSION"

missing() {
  echo "Protobuf compiler (protoc) $REQ_PROTOC_VERSION or later is missing on this system."
  echo "Install protoc ${REQ_PROTOC_VERSION} using the preferred method for your system."
  echo "Refer to https://developers.google.com/protocol-buffers/ if Protobuf $REQ_PROTOC_VERSION is not available in the system repositories."

  exit 1
}

check_protoc_version() {

  INST_PROTOC_VERSION=$(protoc --version | sed -e 's/.*libprotoc *//')
  INST_PROTOC_MAJOR_VERSION=$(echo "$INST_PROTOC_VERSION" | cut -d. -f1)
  INST_PROTOC_MINOR_VERSION=$(echo "$INST_PROTOC_VERSION" | cut -d. -f2)

  if [ "$REQ_PROTOC_MAJOR_VERSION" -gt "$INST_PROTOC_MAJOR_VERSION" ]; then
    missing
  elif [ "$REQ_PROTOC_MAJOR_VERSION" -eq "$INST_PROTOC_MAJOR_VERSION" ] &&
    [ "$REQ_PROTOC_MINOR_VERSION" -gt "$INST_PROTOC_MINOR_VERSION" ]; then
    missing
  fi
}

check_protoc_version

echo "protoc $REQ_PROTOC_VERSION or later is available on the system."
