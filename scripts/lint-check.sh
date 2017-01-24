#/bin/bash

# NOTE: This script needs be run from the root of the GD2 repository

# Find all Go source files in the repository, that are not vendored or generated
# and then run golint on them

RETVAL=0

for file in $(find . -path ./vendor -prune -o -path ./rpc/sunrpcserver -prune -o -type f -name '*.go' -not -name '*.pb.go' -print); do
  golint -set_exit_status $file
  if [ $? -eq 1 -a $RETVAL -eq 0 ]; then
    RETVAL=1
  fi
done
exit $RETVAL
