# Quick Start User Guide

This guide demonstrates creating a **two-node GlusterFS cluster** using glusterd2 and assumes that the user has used GlusterFS before and is familiar with terms such as bricks and volumes.

## Setup

This guide takes the following as an example of IPs for the two nodes:

 * **Node 1**: `192.168.56.101`
 * **Node 2**: `192.168.56.102`

Please follow these steps for setting up glusterd2 on **each of the two nodes**.

### Installing Glusterfs

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

For other distros, you can download binaries of the [latest release](https://github.com/gluster/glusterd2/releases) from Github. 
Like all Go programs, glusterd2 is a single binary (statically linked) without 
external dependencies.

**Config File:** Default path of glusterd2 config file is `/etc/glusterd2/glusterd2.toml`.

### Using external etcd.

Setup etcd on a node. Edit glusterd2 config file by adding `noembed` and `etcdenpoints` options. Replace 
endpoints argument to point to etcd client URL:

```toml
etcdendpoints = "http://[ip_address]:[port]"
noembed = true
```

### Running glusterd2 for RPM installation

**Enable glusterd2 service:** Ensure that glusterd2 service starts automatically when system starts.
To enable glusterd2 service run:

```sh
# systemctl enable glusterd2
```

**Start glusterd2 service:** To start glusterd2 process run:

```sh
# systemctl start glusterd2
```

Check the status of glusterd2 service:

```sh
# systemctl status glusterd2
```
Please ensure that glusterd2 service status is "active (runnning)" before proceeding.


### Running glusterd2 for binaries installation

**Start glusterd2 process:** Create a config file and provide the path of config file while running glusterd2 as shown below.

```sh
# ./glusterd2 --config conf.toml
```

### Authentication

In glusterd2, REST API authentication is enabled by default. To disable rest authentication add `restauth=false` in Glusterd2 config file(`/etc/glusterd2/glusterd2.toml`) or the custom config file provided by you (conf.toml as per above example)

Please restart glusterd2 service after changing config file.

### Using Glusterd2

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
