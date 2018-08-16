#!/bin/bash

. $(dirname $0)/../include.rc
. $(dirname $0)/../volume.rc

cleanup;

TEST daemonize /usr/sbin/glusterd2 --peeraddress="$(hostname):24008" --logdir=/var/log/glusterd2 --logfile=glusterd2.log

sleep 10;

TEST pidof glusterd2
TEST $CLI volume info;

TEST $CLI volume create $V0 replica 2 $H0:$B0/${V0}{1,2,3,4,5,6,7,8} --create-brick-dir --allow-root-dir;

EXPECT "$V0" volinfo_field $V0 'Volume Name';
EXPECT 'Created' volinfo_field $V0 'State';
EXPECT '8' brick_count $V0

TEST $CLI volume start $V0;
EXPECT 'Started' volinfo_field $V0 'State';

TEST $CLI volume add-brick $V0 $H0:$B0/${V0}{9,10,11,12} --create-brick-dir --allow-root-dir;
EXPECT '12' brick_count $V0

# TEST $CLI volume remove-brick $V0 $H0:$B0/${V0}{1,2,3,4} force;
# EXPECT '8' brick_count $V0

TEST $CLI volume stop $V0;
EXPECT 'Stopped' volinfo_field $V0 'State';

TEST $CLI volume delete $V0;
TEST ! $CLI volume info $V0;

cleanup;
