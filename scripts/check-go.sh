#!/bin/bash

# We require a minimum of Go 1.5 as we make use of the vendor/ directory

missing() {
  echo "Go1.5 is missing on this system."
  echo "Install Go${GOVERSION} using the preferred method for your system."
  echo "Refer to https://golang.org/doc/install is Go1.5 is not available in the system repositories."

  exit 1
}

#Check if Go is installed
env go version >/dev/null 2>&1 || missing

# The `link` tool was introduced in Go1.5
# TODO: Do a proper version check. This will be required for gcc-go
if [ ! -e $(env go env GOTOOLDIR)/link ]; then
  missing
fi

echo "Go1.5 is available on the system."
