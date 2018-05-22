#!/bin/bash

PREFIX=${PREFIX:-/usr/local}
BASE_PREFIX=$PREFIX
if [ "$PREFIX" = "/usr" ]; then
    BASE_PREFIX=""
fi

DATADIR=${DATADIR:-$PREFIX/share}
LOCALSTATEDIR=${LOCALSTATEDIR:-$PREFIX/var/lib}
LOGDIR=${LOGDIR:-$BASE_PREFIX/var/log}
RUNDIR=${RUNDIR:-$BASE_PREFIX/var/run}

GD2="glusterd2"
GD2STATEDIR=${GD2STATEDIR:-$LOCALSTATEDIR/$GD2}
GD2LOGDIR=${GD2LOGDIR:-$LOGDIR/$GD2}
GD2RUNDIR=${GD2RUNDIR:-$RUNDIR/$GD2}

OUTDIR=${1:-build}
mkdir -p "$OUTDIR"

OUTPUT=$OUTDIR/$GD2.toml


cat >"$OUTPUT" <<EOF

workdir = "$GD2STATEDIR"
localstatedir = "$GD2STATEDIR"
logdir = "$GD2LOGDIR"
logfile = "$GD2.log"
loglevel = "INFO"
rundir = "$GD2RUNDIR"
defaultpeerport = "24008"
peeraddress = ":24008"
clientaddress = ":24007"
EOF
