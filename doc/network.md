# Network communication and services

This document enumerates various network communication and services present
in Glusterd2 and the network ports used by them. All the port numbers listed
here are configurable via command line option and config file.

## Servers

### REST
The REST server accepts HTTP requests from clients. The HTTP requests sent by
the clients conform to the REST API defined and published by glusterd2.

Default port: 24007

### SunRPC
The SunRPC server accepts TCP connections from:
* Glusterfs clients (FUSE/libgfapi)
* Gluster daemons co-located on same node (bricks and other daemons)

Default port: 24007

Glusterd2 also uses SunRPC to communicate with co-located daemons (bricks etc)
over Unix Domain Sockets.

### gRPC
gRPC is used only for glusterd2 to glusterd2 communication and should not be
exposed to external clients.

Default port: 24008

### etcd
A subset of glusterd2 nodes will have embedded etcd server running. These etcd
server instances talk to each other and also serve etcd clients.

Default ports:
* 2380 for etcd to etcd peer communication
* 2379 for client traffic

## Firewall configuration
Only port `24007` should be exposed to external consumers i.e
* HTTP clients (management ops)
* Glusterfs clients (I/O)

The ports used by gRPC and etcd should be shielded from external network.

## NTP/chronyd
For etcd servers to work reliably, the difference in time between peers in the
cluster should be less than one second. Please configure the NTP service or
manually sync the clocks on different machines.

## RDMA?
Glusterd2 will not support access over RDMA because services offered by it are
not in the I/O path.
