package sunrpc

import (
	"bytes"
	"net"
	"sync/atomic"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/sunrpc"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/rasky/go-xdr/xdr2"
	log "github.com/sirupsen/logrus"
)

// NOTE:
// This adds support for the 'callback hack' from Gluster's RPC implementation
// i.e enable RPC server (glusterd2) to make a call to a connected glusterfs
// SunRPC client.
// This implementation depends on the following facts:
//     1. Multiple goroutines may invoke methods on a net.Conn simultaneously.
//     2. SunRPC ServerCodec will always send entire RPC message in a single
//        RPC fragment/record over the socket.
//     3. The glusterfs RPC client processes will never send a RPC reply to
//        these RPC calls sent by glusterd2.
// If any of the above pre-conditions change, this implementation should be
// revisited.

// Stuff from glusterd1 that uses RPC callbacks:
// - glusterd_fetchspec_notify()
// - glusterd_fetchsnap_notify()
// - glusterd_client_statedump_submit_req()->rpcsvc_request_submit()

// TODO:
// Glusterd2 (and glusterd1) cannot yet recognize clients (as glusterfsd,
// snapd, glusterfs etc). So these callback notifications are sent to all
// the connected RPC clients.

const (
	glusterCbkProgram = 52743234 // GLUSTER_CBK_PROGRAM
	glusterCbkVersion = 1        // GLUSTER_CBK_VERSION
)

var xidCounter uint32

func getNewXid() uint32 {
	return atomic.AddUint32(&xidCounter, 1)
}

func callbackClient(conn net.Conn, p sunrpc.ProcedureID, args interface{}) error {
	payload := new(bytes.Buffer)

	call := sunrpc.RPCMsg{
		Xid:  getNewXid(),
		Type: sunrpc.Call,
		CBody: sunrpc.CallBody{
			RPCVersion: sunrpc.RPCProtocolVersion,
			Program:    p.ProgramNumber,
			Version:    p.ProgramVersion,
			Procedure:  p.ProcedureNumber,
		},
	}

	if _, err := xdr.Marshal(payload, &call); err != nil {
		return err
	}

	if args != nil {
		if _, err := xdr.Marshal(payload, &args); err != nil {
			return err
		}
	}

	_, err := sunrpc.WriteFullRecord(conn, payload.Bytes())
	if err != nil {
		return err
	}

	return nil
}

type fetchOp uint8

// rpc/rpc-lib/src/protocol-common.h:gf_cbk_procnum
const (
	gfCbkFetchSpec fetchOp = 1
	gfCbkGetSnaps  fetchOp = 4
)
const (
	gfCbkStatedump = 9
)

func fetchNotify(t transaction.TxnCtx, op fetchOp) {
	clientsList.RLock()
	defer clientsList.RUnlock()

	p := sunrpc.ProcedureID{
		ProgramNumber:   glusterCbkProgram,
		ProgramVersion:  glusterCbkVersion,
		ProcedureNumber: uint32(op),
	}

	for conn := range clientsList.c {
		go func(c net.Conn) {
			if err := callbackClient(c, p, nil); err != nil {
				t.Logger().WithError(err).WithFields(log.Fields{
					"client":    c.RemoteAddr().String(),
					"procedure": op,
				}).Warn("Failed to notify RPC client")
			}
		}(conn)
	}
}

// FetchSpecNotify notifies all clients connected to glusterd that the volfile
// has changed and the clients should fetch the new volfile.
func FetchSpecNotify(t transaction.TxnCtx) {
	fetchNotify(t, gfCbkFetchSpec)
}

// FetchSnapNotify notifies all clients connected to glusterd that a snapshot
// has been created or modified.
func FetchSnapNotify(t transaction.TxnCtx) {
	fetchNotify(t, gfCbkGetSnaps)
}

type gfStatedump struct {
	Pid uint32
}

// ClientStatedump sends notification to all connected RPC clients on the
// specified host to take statedump. The clients will examine if the PID
// it received in notification is same as it's own PID. If yes, it will
// take it's own statedump.
func ClientStatedump(volname string, host string, pid int, logger log.FieldLogger) {

	clientsList.RLock()
	defer clientsList.RUnlock()

	p := sunrpc.ProcedureID{
		ProgramNumber:   glusterCbkProgram,
		ProgramVersion:  glusterCbkVersion,
		ProcedureNumber: gfCbkStatedump,
	}

	for conn := range clientsList.c {
		h, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		if utils.IsAddressSame(h, host) {
			// TODO: Our RPC framework doesn't yet support tagging
			// each connection with information whether it's a
			// client or brick, or which volume the client is
			// connected to etc. We can filter by volume name here
			// and send RPCs to only those clients.
			go func(c net.Conn) {
				if err := callbackClient(c, p, &gfStatedump{uint32(pid)}); err != nil {
					logger.WithError(err).WithFields(log.Fields{
						"client":    c.RemoteAddr().String(),
						"procedure": gfCbkStatedump,
					}).Warn("Failed to notify RPC client")
				}
			}(conn)
		}
	}
}
