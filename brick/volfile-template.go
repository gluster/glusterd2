package brick

// VolfileTemplate contains bare minimum set of xlators required to start
// brick process.
var VolfileTemplate = `volume <volume-name>-posix
    type storage/posix
    option volume-id <volume-id>
    option directory <brick-path>
end-volume

volume <volume-name>-access-control
    type features/access-control
    subvolumes <volume-name>-posix
end-volume

volume <volume-name>-locks
    type features/locks
    subvolumes <volume-name>-access-control
end-volume

volume <volume-name>-io-threads
    type performance/io-threads
    subvolumes <volume-name>-locks
end-volume

volume <volume-name>-index
    type features/index
    option xattrop-pending-watchlist trusted.afr.<volume-name>-
    option xattrop-dirty-watchlist trusted.afr.dirty
    option index-base <brick-path>/.glusterfs/indices
    subvolumes <volume-name>-io-threads
end-volume

volume <brick-path>
    type performance/decompounder
    subvolumes <volume-name>-index
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
