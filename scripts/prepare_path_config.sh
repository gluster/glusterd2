#!/bin/bash

# Create glusterd2/paths_config.go
glusterd2dir=$1

OUTPUT=${glusterd2dir}/paths_config.go


cat >"$OUTPUT" <<EOF
package main

var defaultConfDir = "${SYSCONFDIR}/glusterd2"
EOF
