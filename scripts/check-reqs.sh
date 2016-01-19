#!/bin/bash

# Checks if the tools required for properly building, verifying and testing GlusterD are installed.

TOOLS=(glide gometalinter etcd)

MISSING=0

for tool in ${TOOLS[@]}; do
  type $tool >/dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "$tool is missing on the system"
    MISSING=1
  else
    echo "$tool is available"
  fi
done

if [ $MISSING -ne 0 ]; then
  exit 1
fi
