# Quick Start User Guide

This guide demonstrates creating a **two-node GlusterFS cluster** using glusterd2 and assumes that the user has used GlusterFS before and is familiar with terms such as bricks and volumes.

## Setup

This guide takes the following as an example of IPs for the two nodes:

 * **Node 1**: `192.168.56.201`
 * **Node 2**: `192.168.56.202`

Please follow these steps for setup on **each of the two nodes**.

### Installing dependent packages

Install rpcbind:

```sh
# dnf install rpcbind
# mkdir -p /run/rpcbind
# systemctl start rpcbind
# systemctl enable rpcbind
```

Install packages that provide GlusterFS server (brick process) and client (fuse, libgfapi):

```sh
# wget -P /etc/yum.repos.d/ https://download.gluster.org/pub/gluster/glusterfs/LATEST/Fedora/glusterfs-fedora.repo
# dnf install glusterfs-server glusterfs-fuse glusterfs-api
```

### Download glusterd2

Glusterd2 is a single binary without any external dependencies. Like all Go programs, dependencies are statically linked. You can download the [latest release](https://github.com/gluster/glusterd2/releases) from Github.

```sh
$ wget https://github.com/gluster/glusterd2/releases/download/v4.0dev-7/glusterd2-v4.0dev-7-linux-amd64.tar.xz
$ tar -xf glusterd2-v4.0dev-7-linux-amd64.tar.xz
```

### Running glusterd2

**Create a working directory:** This is where glusterd2 will store all data which includes logs, pid files, etcd information etc. For this example, we will be using a temporary path. If a working directory is not specified, it defaults to current directory.

```sh
$ mkdir -p /tmp/gd2-workdir
```

**Create a config file:** This is optional but if your VM/machine has multiple network interfaces, it is recommended to create a config file.

```yaml
$ cat conf.yaml 
workdir: "/tmp/gd2-workdir"
peeraddress: "192.168.56.26:24008"
clientaddress: "192.168.56.26:24007"
etcdcurls: "http://192.168.56.26:2379"
etcdpurls: "http://192.168.56.26:2380"
```

Replace the IP address accordingly on each node.

**Start glusterd2 process:** Glusterd2 is not a daemon and currently can run only in the foreground.

```sh
# ./glusterd2 --config conf.yaml
```

You will see an output similar to the following:
```log
INFO[2017-08-28T16:03:58+05:30] Starting GlusterD                             pid=1650
INFO[2017-08-28T16:03:58+05:30] loaded configuration from file                file=conf.yaml
INFO[2017-08-28T16:03:58+05:30] Generated new UUID                            uuid=19db62df-799b-47f1-80e4-0f5400896e05
INFO[2017-08-28T16:03:58+05:30] started muxsrv listener                      
INFO[2017-08-28T16:03:58+05:30] Started GlusterD ReST server                  ip:port=192.168.56.26:24007
INFO[2017-08-28T16:03:58+05:30] Registered RPC Listener                       ip:port=192.168.56.26:24008
INFO[2017-08-28T16:03:58+05:30] started GlusterD SunRPC server                ip:port=192.168.56.26:24007
```

Now you have two nodes running glusterd2.

> NOTE: Ensure that firewalld is configured (or stopped) to let traffic on ports ` before attaching a peer.

### Attach peer

Glusterd2 natively provides only ReST API for clients to perform management operations. A CLI is provided which interacts with glusterd2 using the [ReST APIs](../../wiki/ReST-API).

**Add `node2 (192.168.56.102)` as a peer from `node1 (192.168.56.101)`:**

Create a json file which will contain request body on `node1`:

```sh
$ cat addpeer.json 
{
	"addresses": ["192.168.56.102"]
}
```

Send a HTTP request to `node1` to add `node2` as peer:

```sh
$ curl -X POST http://192.168.56.101:24007/v1/peers --data @addpeer.json -H 'Content-Type: application/json'
```

You will get the Peer ID of the newly added peer as response.

## List peers

Peers in two node cluster can be listed with the following request:

```sh
$ curl -X GET http://192.168.56.101:24007/v1/peers
```

## Create a volume

Create a  JSON file for volume create request body:

```sh
$ cat volcreate.json 
{
	    "name": "testvol",
	    "replica" : 2,
	    "bricks": [
		"192.168.56.101:/export/brick1/data",
		"192.168.56.102:/export/brick2/data",
		"192.168.56.101:/export/brick3/data",
		"192.168.56.102:/export/brick4/data"
	    ],
	    "force": true
}
```

Create brick paths accordingly on each of the two nodes:

 On node1: `mkdir -p /export/brick{1,3}/data`  
 On node2: `mkdir -p /export/brick{2,4}/data`

Send the volume create request to create a 2x2 distributed-replicate volume:

```sh
$ curl -X POST http://192.168.56.101:24007/v1/volumes --data @volcreate.json -H 'Content-Type: application/json'
```

### Start the volume

```sh
$ curl -X POST http://192.168.56.101:24007/v1/volumes/testvol/start
```

Verify that `glusterfsd` process is running on both nodes.

### Mount the volume

```sh
#  mount -t glusterfs 192.168.56.101:testvol /mnt
```

> NOTE: IP of any of the two nodes can be used by ReST clients and mount clients.

### Known issues

* Restarting glusterd2 does not restore the cluster
* Peer detach doesn't work on a 2 node cluster
