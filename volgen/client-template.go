package volgen

var clientVolfileBaseTemplate = `
volume <volume-name>-write-behind
    type performance/write-behind
    subvolumes <volume-name>-<wb-subvol>
end-volume

volume <volume-name>-read-ahead
    type performance/read-ahead
    subvolumes <volume-name>-write-behind
end-volume

volume <volume-name>-readdir-ahead
    type performance/readdir-ahead
    subvolumes <volume-name>-read-ahead
end-volume

volume <volume-name>-io-cache
    type performance/io-cache
    subvolumes <volume-name>-readdir-ahead
end-volume

volume <volume-name>-quick-read
    type performance/quick-read
    subvolumes <volume-name>-io-cache
end-volume

volume <volume-name>-open-behind
    type performance/open-behind
    subvolumes <volume-name>-quick-read
end-volume

volume <volume-name>-md-cache
    type performance/md-cache
    subvolumes <volume-name>-open-behind
end-volume

volume <volume-name>-io-threads
    type performance/io-threads
    subvolumes <volume-name>-md-cache
end-volume

volume <volume-name>
    type debug/io-stats
    option count-fop-hits off
    option latency-measurement off
    option log-level INFO
    subvolumes <volume-name>-io-threads
end-volume
`

var clientVolfileDHTTemplate = `
volume <volume-name>-dht
    type cluster/distribute
    option lock-migration off
    subvolumes <dht-subvolumes>
end-volume
`

var clientVolfileAFRTemplate = `
volume <volume-name>-replicate<child-index>
    type cluster/replicate
    option use-compound-fops off
    option afr-pending-xattr <afr-pending-xattr>
    subvolumes <afr-subvolumes>
end-volume
`

var clientLeafTemplate = `
volume <volume-name>-client-<child-index>
    type protocol/client
    option send-gids true
    option password <trusted-password>
    option username <trusted-username>
    option transport.address-family inet
    option transport-type tcp
    option remote-subvolume <brick-path>
    option remote-host <remote-host>
    option ping-timeout 42
end-volume
`
