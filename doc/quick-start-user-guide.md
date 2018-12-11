# Quick Start User Guide

This guide demonstrates creating a **two-node GlusterFS cluster** using glusterd2 and assumes that the user has used GlusterFS before and is familiar with terms such as bricks and volumes.

## Setup

This guide takes the following as an example of IPs for the two nodes:

 * **Node 1**: `192.168.56.101`
 * **Node 2**: `192.168.56.102`

Please follow these steps for setup on **each of the two nodes**.

### Installing dependent packages

Install rpcbind:

```sh
# dnf install rpcbind
# mkdir -p /run/rpcbind
# systemctl start rpcbind
# systemctl enable rpcbind
```

> **IMPORTANT:** Please install glusterfs from source using code from the [master branch](https://github.com/gluster/glusterfs/tree/master) OR if on CentOS 7, you can install glusterfs using nightly RPMs.

**Installing glusterfs from nightly RPMs (CentOS 7):**

These packages require dependencies present in [EPEL](https://fedoraproject.org/wiki/EPEL). Enable the [EPEL repositories](https://fedoraproject.org/wiki/EPEL#Quickstart) before enabling gluster nightly packages repo below.

Install packages that provide GlusterFS server (brick process) and client (fuse, libgfapi):

```sh
# curl -o /etc/yum.repos.d/glusterfs-nightly-master.repo http://artifacts.ci.centos.org/gluster/nightly/master.repo
# yum install glusterfs-server glusterfs-fuse glusterfs-api
```

### Download glusterd2

We recommend that you use the RPMs made available nightly as they contain the latest fixes. If you are on CentOS 7, you can download the latest glusterd2 nightly RPM using the following method:

```sh
# curl -o /etc/yum.repos.d/glusterd2-nightly-master.repo http://artifacts.ci.centos.org/gluster/gd2-nightly/gd2-master.repo
# yum install glusterd2
```

Alternatlively, if you are using a non-RPM based distro, you can download
binaries of the latest release. Like all Go programs, glusterd2 is a single
binary (statically linked) without external dependencies. You can download the
[latest release](https://github.com/gluster/glusterd2/releases) from Github.

### Running glusterd2

**Create a working directory:** This is where glusterd2 will store all data which includes logs, pid files, etcd information etc. For this example, we will be using a temporary path. If a working directory is not specified, it defaults to current directory.

```sh
$ mkdir -p /var/lib/gd2
```

**Create a config file:** This is optional but if your VM/machine has multiple network interfaces, it is recommended to create a config file. The config file location can be passed to Glusterd2 using the `--config` option.
Glusterd2 will also pick up conf file named `glusterd2.toml` if available in `/etc/glusterd2/` or the current directory.

```toml
$ cat conf.toml
localstatedir = "/var/lib/gd2"
peeraddress = "192.168.56.101:24008"
clientaddress = "192.168.56.101:24007"
etcdcurls = "http://192.168.56.101:2379"
etcdpurls = "http://192.168.56.101:2380"
```

Replace the IP address accordingly on each node.

**Start glusterd2 process:** Glusterd2 is not a daemon and currently can run only in the foreground.

```sh
# ./glusterd2 --config conf.toml
```

You will see an output similar to the following:
```log
INFO[2017-08-28T16:03:58+05:30] Starting GlusterD                             pid=1650
INFO[2017-08-28T16:03:58+05:30] loaded configuration from file                file=conf.toml
INFO[2017-08-28T16:03:58+05:30] Generated new UUID                            uuid=19db62df-799b-47f1-80e4-0f5400896e05
INFO[2017-08-28T16:03:58+05:30] started muxsrv listener                      
INFO[2017-08-28T16:03:58+05:30] Started GlusterD ReST server                  ip:port=192.168.56.101:24007
INFO[2017-08-28T16:03:58+05:30] Registered RPC Listener                       ip:port=192.168.56.101:24008
INFO[2017-08-28T16:03:58+05:30] started GlusterD SunRPC server                ip:port=192.168.56.101:24007
```

Now you have two nodes running glusterd2.

> NOTE: Ensure that firewalld is configured (or stopped) to let traffic on ports ` before adding a peer.

## Add peer

Glusterd2 natively provides only ReST API for clients to perform management operations. A CLI is provided which interacts with glusterd2 using the [ReST APIs](https://github.com/gluster/glusterd2/wiki/ReST-API).

**Add `node2 (192.168.56.102)` as a peer from `node1 (192.168.56.101)`:**

Create a json file which will contain request body on `node1`:

```sh
$ cat addpeer.json 
{
	"addresses": ["192.168.56.102"]
}
```
`addresses` takes a list of address by which the new host can be added. It can be FQDNs, short-names or IP addresses. Note that if you want to add multiple peers use below API to add each peer one at a time.

Send a HTTP request to `node1` to add `node2` as peer:

```sh
$ curl -X POST http://192.168.56.101:24007/v1/peers --data @addpeer.json -H 'Content-Type: application/json'
```

or using glustercli:

    $ glustercli peer add 192.168.56.102

You will get the Peer ID of the newly added peer as response.

## List peers

Peers in two node cluster can be listed with the following request:

```sh
$ curl -X GET http://192.168.56.101:24007/v1/peers
```

or by using the glustercli:

    $ glustercli peer list

Note the UUIDs in the response. We will use the same in volume create request below.

## Create a volume

Create a  JSON file for volume create request body:

```sh
$ cat volcreate.json
{
        "name": "testvol",
        "subvols": [
            {
                "type": "replicate",
                "bricks": [
                    {"peerid": "<uuid1>", "path": "/export/brick1/data"},
                    {"peerid": "<uuid2>", "path": "/export/brick2/data"}
                ],
                "replica": 2
            },
            {
                "type": "replicate",
                "bricks": [
                    {"peerid": "<uuid1>", "path": "/export/brick3/data"},
                    {"peerid": "<uuid2>", "path": "/export/brick4/data"}
                ],
                "replica": 2
            }
        ],
        "force": true
}
```

Insert the actual UUID of the two glusterd2 instances in the above json file.

Create brick paths accordingly on each of the two nodes:

 On node1: `mkdir -p /export/brick{1,3}/data`
 On node2: `mkdir -p /export/brick{2,4}/data`

Send the volume create request to create a 2x2 distributed-replicate volume:

```sh
$ curl -X POST http://192.168.56.101:24007/v1/volumes --data @volcreate.json -H 'Content-Type: application/json'
```

Send the volume create request using glustercli:

    $ glustercli volume create --name testvol <uuid1>:/export/brick1/data <uuid2>:/export/brick2/data <uuid1>:/export/brick3/data <uuid2>:/export/brick4/data --replica 2

## Start the volume

Send the volume start request:

```sh
$ curl -X POST http://192.168.56.101:24007/v1/volumes/testvol/start
```
 or using glustercli:

     $ glustercli volume start testvol

Verify that `glusterfsd` process is running on both nodes.

## Mount the volume

```sh
#  mount -t glusterfs 192.168.56.101:testvol /mnt
```

> NOTE: IP of any of the two nodes can be used by ReST clients and mount clients.

### Known issues

* Issues with 2 node clusters
  * Restarting glusterd2 does not restore the cluster
  * Peer remove doesn't work
