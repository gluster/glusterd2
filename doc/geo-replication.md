# Managing Geo-replication

Geo-replication provides a continuous, asynchronous, and incremental
replication service from one Gluster Volume to another over Local Area
Networks(LANs), Wide Area Network (WANs), and across the Internet.

## Replicated Volumes vs Geo-replication

The following table lists the difference between replicated volumes
and geo-replication:

| Replicated Volumes | Geo-replication |
| -------------------| --------------- |
|Mirrors data across clusters | Mirrors data across geographically distributed clusters |
|Provides high-availability	  | Ensures backing up of data for disaster recovery        |
|Synchronous replication (each and every file operation is sent across all the bricks)	| Asynchronous replication (checks for the changes in files periodically and syncs them on detecting differences) |

## Prerequisites

### Time Synchronization

On bricks of a geo-replication master volume, all the servers' time
must be uniform. You are recommended to set up NTP (Network Time
Protocol) service to keep the bricks sync in time and avoid
out-of-time sync effect.

For example: In a Replicated volume where brick1 of the master is at
12.20 hrs and brick 2 of the master is at 12.10 hrs with 10 minutes
time lag, all the changes in brick2 between this period may go
unnoticed during synchronization of files with remote volume.

### Master and Remote Volumes

Geo-replication will not create remote volume automatically. To
establish a Geo-replication session make sure Master and Remote
volumes are in started state.

**Note**: Glusterd2 will not recognise the volumes created using older
version of glusterd. Master and Remote volumes should be created using
glusterd2.

### Firewall, Ports and Remote addresses

- During session create, Geo-replication needs access to remote glusterd2
  port(Default is `24007`).
- Once the session is established, Geo-replication workers(one per
  Master brick) will use `ssh` to sync files.
- Once the session is established, Master workers will get the remote
  IP/hostnames from remote volume's info. If the internal IPs are used
  to create remote volume then Geo-rep workers will not be able to
  sync the files.

## Creating the session - Master and Remote volumes are in same cluster

Create a geo-rep session between master and remote volume using the
following command.

```
# glustercli geo-replication create <master-volume-name> \
    [<remote-user>@]<remote-host>::<remote-volume-name> [--force]
```

For example,

```
# glustercli geo-replication create gv1 root@rnode1.example.com::gv2
```

## Creating the session - Master and Remote volumes are in different cluster

If remote volume is part of different cluster than Master volume, use
`--remote-endpoints` to specify the remote credentials.

For example,

```
# glustercli geo-replication create gv1 root@rnode1.example.com::gv2 \
    --remote-endpoints=http://rnode1.example.com:24007 \
    --remote-secret=mysecret
```

## Starting Geo-replication

Use the following command to start the geo-replication session,

```
# glustercli geo-replication start <master-volume-name> \
    [<remote-user>@]<remote-host>::<remote-volume-name>
```

For example,

```
# glustercli geo-replication start gv1 root@rnode1.example.com::gv2
```

## Stopping Geo-replication

Use the following command to stop the geo-replication session,

```
# glustercli geo-replication stop <master-volume-name> \
    [<remote-user>@]<remote-host>::<remote-volume-name>
```

For example,

```
# glustercli geo-replication stop gv1 root@rnode1.example.com::gv2
```

## Status

Geo-replication session status can be checked by running the status
command in one of the node in Master Cluster.

To check the status of all Geo-replication sessions in the Cluster

```
# glustercli geo-replication status
```

To check the status of one session,

```
# glustercli geo-replication status <master-volume-name> \
    [<remote-user>@]<remote-host>::<remote-volume-name>
```

Example,

```
# glustercli geo-replication status gv1 root@rnode1::gv2
```

Example Status output,

```
SESSION: gv1 ==> root@gluster1.redhat.com::gv2  STATUS: Started
+----------------------------------------------+--------+-----------------+--------------------+---------------------+-----------------+----------------------------+
|                 MASTER BRICK                 | STATUS |  CRAWL STATUS   | REMOTE NODE        |     LAST SYNCED     | CHECKPOINT TIME | CHECKPOINT COMPLETION TIME |
+----------------------------------------------+--------+-----------------+--------------------+---------------------+-----------------+----------------------------+
| mnode1.example.com:/bricks/gv1/brick1/brick  | Active | Changelog Crawl | snode1.example.com | 2018-07-21 16:03:51 | N/A             | N/A                        |
| mnode2.example.com:/bricks/gv1/brick2/brick  | Active | Changelog Crawl | snode2.example.com | 2018-07-21 15:55:36 | N/A             | N/A                        |
| mnode3.example.com:/bricks/gv1/brick3/brick  | Active | Changelog Crawl | snode3.example.com | 2018-07-21 16:01:23 | N/A             | N/A                        |
+----------------------------------------------+--------+-----------------+--------------------+---------------------+-----------------+----------------------------+
```

The STATUS of the session could be one of the following,

- **Initializing**: This is the initial phase of the Geo-replication
  session; it remains in this state for a minute in order to make sure
  no abnormalities are present.
- **Created**: The geo-replication session is created, but not started.
- **Active**: The gsync daemon in this node is active and syncing the
  data.(One worker among the replica pairs will be in Active state)
- **Passive**: A replica pair of the active node. The data
  synchronization is handled by active node. Hence, this node does not
  sync any data. If Active node goes down, Passive worker will become
  Active
- **Faulty**: The geo-replication session has experienced a problem,
  and the issue needs to be investigated further. Check log files for
  more details about the Faulty status.
- **Stopped**: The geo-replication session has stopped, but has not
  been deleted.

The CRAWL STATUS can be one of the following:

- **Hybrid Crawl**: The gsyncd daemon is crawling the glusterFS file
  system and generating pseudo changelog to sync data. This crawl is
  used during initial sync and if Changelogs are not available.
- **History Crawl**: gsyncd daemon syncs data by consuming Historical
  Changelogs. On every worker restart, Geo-rep uses this Crawl to
  process backlog Changelogs.
- **Changelog Crawl**: The changelog translator has produced the
  changelog and that is being consumed by gsyncd daemon to sync data.

## Checkpoint

Using Checkpoint feature we can find the status of sync with respect
to the Checkpoint time. Checkpoint completion status shows "Yes" once
Geo-rep syncs all the data from that brick which are created or
modified before the Checkpoint Time.

Set the Checkpoint using,

```
# glustercli geo-replication set <master-volume-name> \
    [<remote-user>@]<remote-host>::<remote-volume-name> \
    checkpoint now
```

Example,

```
# glustercli geo-replication set gv1 root@rnode1.example.com::gv2 \
    checkpoint now
```

Touch the Master mount point to make sure Checkpoint completes even
though no I/O happening in the Volume

```
# mount -t glusterfs <masterhost>:<mastervol> /mnt
# touch /mnt
```

Checkpoint status can be checked using Geo-rep status
command. Following columns in status output gives more information
about Checkpoint

- **CHECKPOINT TIME**: Checkpoint Set Time
- **CHECKPOINT COMPLETED**: Yes/No/NA, Status of Checkpoint
- **CHECKPOINT COMPLETION TIME**: Checkpoint Completion Time if
  completed, else N/A


## Deleting the session

Established Geo-replication session can be deleted using the following
command,

```
# glustercli geo-replication delete <master-volume-name> \
    [<remote-user>@]<remote-host>::<remote-volume-name>
```

For example,

```
# glustercli geo-replication delete gv1 root@rnode1.example.com::gv2
```

**Note**: The syncing will resume from where it was stopped before
deleting the session if the same session is created again. The
session can be deleted permanently by using `--reset-sync-time` option
with delete command. For example,

```
# glustercli geo-replication delete gv1 root@rnode1.example.com::gv2 \
    --reset-sync-time
```

## Log Files

- Master Log files are located in `/var/log/glusterfs/geo-replication`
 directory in each master nodes.
- Remote log files are located in
 `/var/log/glusterfs/geo-replication-slaves` directory in remote nodes.
