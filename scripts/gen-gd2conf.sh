#!/bin/bash

OUTDIR=${1:-build}
mkdir -p "$OUTDIR"

OUTPUT=$OUTDIR/$GD2.toml


cat >"$OUTPUT" <<EOF

localstatedir = "$GD2STATEDIR"
logdir = "$GD2LOGDIR"
logfile = "$GD2.log"
loglevel = "INFO"
rundir = "$GD2RUNDIR"
defaultpeerport = "24008"
peeraddress = ":24008"
clientaddress = ":24007"
block-hosting-volume-size = 5368709120 #5GiB
auto-create-block-hosting-volumes = true
block-hosting-volume-replica-count = 3
#restauth should be set to false to disable REST authentication in glusterd2
#restauth = false
EOF
