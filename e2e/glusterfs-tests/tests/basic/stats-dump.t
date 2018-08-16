#!/bin/bash

. $(dirname $0)/../include.rc
. $(dirname $0)/../volume.rc
. $(dirname $0)/../nfs.rc

cleanup;

TEST glusterd
TEST pidof glusterd
TEST $CLI volume create $V0 replica 3 $H0:$B0/${V0}{0,1,2}
TEST $CLI volume set $V0 diagnostics.latency-measurement on
TEST $CLI volume set $V0 diagnostics.count-fop-hits on
TEST $CLI volume set $V0 diagnostics.stats-dump-interval 1
TEST $CLI volume set $V0 performance.nfs.io-threads on
TEST $CLI volume set $V0 nfs.disable off
TEST $CLI volume start $V0
EXPECT_WITHIN $NFS_EXPORT_TIMEOUT "1" is_nfs_export_available
TEST glusterfs --volfile-id=/$V0 --volfile-server=$H0 $M0 --attribute-timeout=0 --entry-timeout=0
TEST mount_nfs $H0:/$V0 $N0 nolock,soft,intr

for i in {1..10};do
  dd if=/dev/zero of=$M0/fuse_testfile$i bs=4k count=100
done

for i in {1..10};do
  dd if=/dev/zero of=$N0/nfs_testfile$i bs=4k count=100
done
sleep 2

# Verify we have non-zero write counts from the bricks, gNFSd
# and the FUSE mount.
TEST [ $(grep 'aggr.fop.write.count' ${GLUSTERD_WORKDIR}/stats/glusterfs_nfsd.dump|tail -1|cut -d: -f2) != "0," ]
TEST [ $(grep 'aggr.fop.write.count' ${GLUSTERD_WORKDIR}/stats/glusterfs_patchy.dump|tail -1|cut -d: -f2) != "0," ]
TEST [ $(grep 'aggr.fop.write.count' ${GLUSTERD_WORKDIR}/stats/glusterfsd__d_backends_patchy0.dump|tail -1|cut -d: -f2) != "0," ]
TEST [ $(grep 'aggr.fop.write.count' ${GLUSTERD_WORKDIR}/stats/glusterfsd__d_backends_patchy1.dump|tail -1|cut -d: -f2) != "0," ]
TEST [ $(grep 'aggr.fop.write.count' ${GLUSTERD_WORKDIR}/stats/glusterfsd__d_backends_patchy2.dump|tail -1|cut -d: -f2) != "0," ]

# Test that io-stats is getting queue sizes from io-threads
TEST grep '.queue_size' ${GLUSTERD_WORKDIR}/stats/glusterfs_nfsd.dump
TEST grep '.queue_size' ${GLUSTERD_WORKDIR}/stats/glusterfsd__d_backends_patchy0.dump
TEST grep '.queue_size' ${GLUSTERD_WORKDIR}/stats/glusterfsd__d_backends_patchy1.dump
TEST grep '.queue_size' ${GLUSTERD_WORKDIR}/stats/glusterfsd__d_backends_patchy2.dump

cleanup;
