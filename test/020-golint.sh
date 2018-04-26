#!/bin/bash

# NOTE: This script needs be run from the root of the GD2 repository

# Find all Go source files in the repository, that are not vendored or generated
# and then run golint on them

RETVAL=0
GENERATED_FILES="*(.pb|_string).go"

for file in $(find . -path ./vendor -prune -o -type f -name '*.go' -print | grep -E -v "$GENERATED_FILES"); do
  golint -set_exit_status "$file"
  if [[ $? -eq 1 ]]; then
    RETVAL=1
  fi
done
exit $RETVAL
