Tracking resource usage in Glusterd2
====================================

Tracing and profiling with `pprof` is intended for developers only. This can be
used for debugging memory leaks/consumption, slow flows through the code and so
on. Users will normally not want profiling enabled on their production systems.
To investigate memory allocations in Glusterd2, it is needed to enable the
profiling feature. This can be done by adding `"profiling": true` in
the `--config` file.
Enabling profiling makes standard Golang pprof endpoints available. For memory
allocations `/debug/pprof/heap` is most useful.
Capturing a snapshot of the current allocations in the Glusterd2 is pretty
simple. On the node running Glusterd2, the go pprof tool command can be used:
```
[root@fedora3 glusterd2]# go tool pprof http://localhost:24007/debug/pprof/heap
File: glusterd2
Build ID: 7a94c2e498445577aaf7f910d6ef1c3adc19d553
Type: inuse_space
Time: Nov 28, 2018 at 2:55pm (IST)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof)
```

For working with the size of the allocations, it is helpful to set all sizes to
`megabytes`, otherwise `auto` is used as a `unit` and that would add human
readable B, kB, MB postfixes. This however is not useful for sorting with
scripts. So, set the `unit` to `megabytes` instead:

```
(pprof) unit=megabytes
```

```
(pprof) top
Showing nodes accounting for 330.75MB, 98.52% of 335.73MB total
Dropped 305 nodes (cum <= 1.68MB)
Showing top 10 nodes out of 14
      flat  flat%   sum%        cum   cum%
     326MB 97.10% 97.10%      326MB 97.10%  github.com/gluster/glusterd2/pkg/sunrpc.ReadFullRecord /root/work/src/github.com/gluster/glusterd2/pkg/sunrpc/record.go
    2.38MB  0.71% 97.81%     2.38MB  0.71%  github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/rafthttp.startStreamWriter /root/work/src/github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/rafthttp/stream.go
    2.38MB  0.71% 98.52%     4.77MB  1.42%  github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/rafthttp.startPeer /root/work/src/github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/rafthttp/peer.go
         0     0% 98.52%   326.01MB 97.10%  github.com/gluster/glusterd2/pkg/sunrpc.(*serverCodec).ReadRequestHeader /root/work/src/github.com/gluster/glusterd2/pkg/sunrpc/servercodec.go
         0     0% 98.52%     4.83MB  1.44%  github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver.(*EtcdServer).apply /root/work/src/github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver/server.go
         0     0% 98.52%     4.83MB  1.44%  github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver.(*EtcdServer).applyAll /root/work/src/github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver/server.go
         0     0% 98.52%     4.77MB  1.42%  github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver.(*EtcdServer).applyConfChange /root/work/src/github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver/server.go
         0     0% 98.52%     4.83MB  1.44%  github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver.(*EtcdServer).applyEntries /root/work/src/github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver/server.go
         0     0% 98.52%     4.83MB  1.44%  github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver.(*EtcdServer).run.func6 /root/work/src/github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/etcdserver/server.go
         0     0% 98.52%     4.83MB  1.44%  github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/pkg/schedule.(*fifo).run /root/work/src/github.com/gluster/glusterd2/vendor/github.com/coreos/etcd/pkg/schedule/schedule.go
(pprof)
```

Looking at the above output, we can see that the first line consumes hundreds
of MB while the other lines are just a fraction of it.
This makes sure that ReadFullRecord is consuming more memory than required.
To understand things even more clearly, we can see which line of ReadFullRecord
consumes the memory. To view that use the list command:

```
(pprof) list ReadFullRecord
Total: 335.73MB
ROUTINE ======================== github.com/gluster/glusterd2/pkg/sunrpc.ReadFullRecord in /root/work/src/github.com/gluster/glusterd2/pkg/sunrpc/clientcodec.go
    8.11kB     8.11kB (flat, cum) 0.0024% of Total
         .          .      1:package sunrpc
         .          .      2:
         .          .      3:import (
    8.11kB     8.11kB      4:	"bytes"
         .          .      5:	"io"
         .          .      6:	"net"
         .          .      7:	"net/rpc"
         .          .      8:	"sync"
         .          .      9:
ROUTINE ======================== github.com/gluster/glusterd2/pkg/sunrpc.ReadFullRecord in /root/work/src/github.com/gluster/glusterd2/pkg/sunrpc/record.go
     326MB      326MB (flat, cum) 97.10% of Total
         .          .     96:func ReadFullRecord(conn io.Reader) ([]byte, error) {
         .          .     97:
         .          .     98:	// In almost all cases, RPC message contain only one fragment which
         .          .     99:	// is not too big in size. But set a cap on buffer size to prevent
         .          .    100:	// rogue clients from filling up memory.
     326MB      326MB    101:	record := bytes.NewBuffer(make([]byte, 0, maxRecordSize))
         .          .    102:	var fragmentHeader uint32
         .          .    103:	for {
         .          .    104:		// Read record fragment header
         .          .    105:		err := binary.Read(conn, binary.BigEndian, &fragmentHeader)
         .          .    106:		if err != nil {
(pprof)
```

The records allocation is the culprit and we need to look further into
the code as to why the records is allocating so much and fix it.
