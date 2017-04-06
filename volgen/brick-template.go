package volgen

var brickVolfileTemplate = `volume <volume-name>-posix
    type storage/posix
    option volume-id <volume-id>
    option directory <brick-path>
end-volume

volume <volume-name>-trash
    type features/trash
    option trash-internal-op off
    option brick-path <brick-path>
    option trash-dir .trashcan
    subvolumes <volume-name>-posix
end-volume

volume <volume-name>-changetimerecorder
    type features/changetimerecorder
    option sql-db-wal-autocheckpoint 25000
    option sql-db-cachesize 12500
    option ctr-record-metadata-heat off
    option record-counters off
    option ctr-enabled off
    option record-entry on
    option ctr_lookupheal_inode_timeout 300
    option ctr_lookupheal_link_timeout 300
    option ctr_link_consistency off
    option record-exit off
    option db-path <brick-path>/.glusterfs/
    option db-name data.db
    option hot-brick off
    option db-type sqlite3
    subvolumes <volume-name>-trash
end-volume

volume <volume-name>-changelog
    type features/changelog
    option changelog-barrier-timeout 120
    option changelog-dir <brick-path>/.glusterfs/changelogs
    option changelog-brick <brick-path>
    subvolumes <volume-name>-changetimerecorder
end-volume

volume <volume-name>-bitrot-stub
    type features/bitrot-stub
    option export <brick-path>
    subvolumes <volume-name>-changelog
end-volume

volume <volume-name>-access-control
    type features/access-control
    subvolumes <volume-name>-bitrot-stub
end-volume

volume <volume-name>-locks
    type features/locks
    subvolumes <volume-name>-access-control
end-volume

volume <volume-name>-worm
    type features/worm
    option worm-file-level off
    option worm off
    subvolumes <volume-name>-locks
end-volume

volume <volume-name>-read-only
    type features/read-only
    option read-only off
    subvolumes <volume-name>-worm
end-volume

volume <volume-name>-leases
    type features/leases
    option leases off
    subvolumes <volume-name>-read-only
end-volume

volume <volume-name>-upcall
    type features/upcall
    option cache-invalidation off
    subvolumes <volume-name>-leases
end-volume

volume <volume-name>-io-threads
    type performance/io-threads
    subvolumes <volume-name>-upcall
end-volume

volume <volume-name>-marker
    type features/marker
    option inode-quota off
    option quota off
    option gsync-force-xtime off
    option xtime off
    option quota-version 0
    option timestamp-file <local-state-dir>/vols/<volume-name>/marker.tstamp
    option volume-uuid <volume-id>
    subvolumes <volume-name>-io-threads
end-volume

volume <volume-name>-barrier
    type features/barrier
    option barrier-timeout 120
    option barrier disable
    subvolumes <volume-name>-marker
end-volume

volume <volume-name>-index
    type features/index
    option xattrop-pending-watchlist trusted.afr.<volume-name>-
    option xattrop-dirty-watchlist trusted.afr.dirty
    option index-base <brick-path>/.glusterfs/indices
    subvolumes <volume-name>-barrier
end-volume

volume <volume-name>-quota
    type features/quota
    option deem-statfs off
    option timeout 0
    option server-quota off
    option volume-uuid <volume-name>
    subvolumes <volume-name>-index
end-volume

volume <volume-name>-io-stats
    type debug/io-stats
    option count-fop-hits off
    option latency-measurement off
    option log-level INFO
    option unique-id <brick-path>
    subvolumes <volume-name>-quota
end-volume

volume <brick-path>
    type performance/decompounder
    subvolumes <volume-name>-io-stats
end-volume

volume <volume-name>-server
    type protocol/server
    option auth.addr.<brick-path>.allow *
    option auth-path <brick-path>
    option auth.login.<trusted-username>.password <trusted-password>
    option auth.login.<brick-path>.allow <trusted-username>
    option transport.address-family inet
    option transport-type tcp
    subvolumes <brick-path>
end-volume
`
