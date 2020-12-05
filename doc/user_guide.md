# User guide:

## Glusterd2 and glusterfs

This helps in understanding how glusterd2 (GD2) goes along with glusterfs.

### Glusterfs

GlusterFS is a scalable distributed network filesystem. More about gluster can be found [here] (https://docs.gluster.org/en/latest/).
***Note:*** An understanding of Glusterfs is necessary to use Glusterd2.

#### Glusterd

Glusterd is the management daemon for glusterfs. The glusterd serves as the Gluster elastic volume manager, overseeing glusterfs processes, and co-ordinating dynamic volume operations, such as adding and removing volumes across multiple storage servers non-disruptively.

Glusterd runs on all the servers. Commands are issued to glusterd using the cli which is a part of glusterd (can be issued on any server running glusterd).

#### Glusterd2

Glusterd2 is the next version of glusterd and its a maintained as a separate project for now.
It works along with glusterfs binaries and more about it will be explained in the installation.

Glusterd2 has its own cli which is different from glusterds'cli.

**Note:** There are other ways to communicate with glusterd2 which is explained in the architecture as well as the [configuring GD2]() section

## Installation

Note: Glusterd and gluster cli (the first version) are installed with the glusterfs. Glusterd2 has to be installed separately as of now.

## Configuring GD2

## Using GD2

### Basics Tasks

[Starting and stopping GD2](doc/managing-the-glusterd2-service.md)
[Managing Trusted Storage Pools](doc/managing-trusted-storage-pool.md)
[Setting Up Storage](https://docs.gluster.org/en/latest/Administrator%20Guide/setting-up-storage/)
[Setting Up Volumes](doc/setting-up-volumes.md)
[Setting Up Clients](https://docs.gluster.org/en/latest/Administrator%20Guide/Setting%20Up%20Clients/)
[Managing GlusterFS Volumes](doc/managing-volumes.md)

### Features

[Geo-replication](doc/geo-replication.md)
[Snapshot](doc/snapshot.md)
[Bit-rot](doc/bitrot.md)
[Quota](doc/quota.md)


## Known Issues

**IMPORTANT:** Do not use glusterd and glusterd2 together. Do not file bugs when done so.

[Known issues](doc/known-issues.md)

## Trouble shooting
