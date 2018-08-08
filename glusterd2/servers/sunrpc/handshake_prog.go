package sunrpc

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc/dict"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/sunrpc"
	"github.com/gluster/glusterd2/plugins/rebalance"

	log "github.com/sirupsen/logrus"
)

const (
	hndskProgNum     = 14398633 // GLUSTER_HNDSK_PROGRAM
	hndskProgVersion = 2        // GLUSTER_HNDSK_VERSION
)

const (
	gfHndskGetSpec       = 2 // GF_HNDSK_GETSPEC
	gfHndskEventNotify   = 5 // GF_HNDSK_EVENT_NOTIFY,
	gfHndskGetVolumeInfo = 6 // GF_HNDSK_GET_VOLUME_INFO

)
const (
	gfEventNotifyDefragStatus = 0
)

var volfilePrefix = "volfiles/"

// GfHandshake is a type for GlusterFS Handshake RPC program
type GfHandshake genericProgram

func newGfHandshake() *GfHandshake {
	// rpc/rpc-lib/src/protocol-common.h
	return &GfHandshake{
		name:        "Gluster Handshake",
		progNum:     hndskProgNum,
		progVersion: hndskProgVersion,
		procedures: []sunrpc.Procedure{
			{ID: sunrpc.ProcedureID{ProgramNumber: hndskProgNum, ProgramVersion: hndskProgVersion,
				ProcedureNumber: gfHndskGetSpec}, Name: "ServerGetspec"},
			{ID: sunrpc.ProcedureID{ProgramNumber: hndskProgNum, ProgramVersion: hndskProgVersion,
				ProcedureNumber: gfHndskGetVolumeInfo}, Name: "ServerGetVolumeInfo"},
			{ID: sunrpc.ProcedureID{ProgramNumber: hndskProgNum, ProgramVersion: hndskProgVersion,
				ProcedureNumber: gfHndskEventNotify}, Name: "ServerEventNotify"},
		},
	}
}

// Name returns the name of the RPC program
func (p *GfHandshake) Name() string {
	return p.name
}

// Number returns the RPC Program number
func (p *GfHandshake) Number() uint32 {
	return p.progNum
}

// Version returns the RPC program version number
func (p *GfHandshake) Version() uint32 {
	return p.progVersion
}

// Procedures returns a list of procedures provided by the RPC program
func (p *GfHandshake) Procedures() []sunrpc.Procedure {
	return p.procedures
}

// GfGetspecReq is sent by glusterfs client and primarily contains volume name.
// Xdata field is a serialized gluster dict containing op version.
type GfGetspecReq struct {
	Flags uint
	Key   string // volume name
	Xdata []byte // serialized dict
}

// GfGetspecRsp is response sent to glusterfs client in response to a
// GfGetspecReq request
type GfGetspecRsp struct {
	OpRet   int
	OpErrno int
	Spec    string // volfile contents
	Xdata   []byte // serialized dict
}

// gf_getspec_flags_type from rpc/rpc-lib/src/protocol-common.h
const (
	gfGetspecFlagServersList = 1
)

// ServerGetspec returns the content of client volfile for the volume
// specified by the client
func (p *GfHandshake) ServerGetspec(args *GfGetspecReq, reply *GfGetspecRsp) error {
	var (
		err      error
		addrs    []string
		respDict map[string]string
	)

	_, err = dict.Unserialize(args.Xdata)
	if err != nil {
		log.WithError(err).Error("ServerGetspec(): dict.Unserialize() failed")
	}

	// Get Volfile from store
	volfileID := strings.TrimPrefix(args.Key, "/")
	resp, err := store.Get(context.TODO(), volfilePrefix+volfileID)
	if err != nil {
		log.WithError(err).WithField("volfile", args.Key).Error("ServerGetspec(): failed to retrive volfile from store")
		goto Out
	}

	if resp.Count != 1 {
		err = errors.New("volfile not found in store")
		log.WithError(err).WithField("volfile", args.Key)
		goto Out
	}

	reply.Spec = string(resp.Kvs[0].Value)
	reply.OpRet = len(reply.Spec)
	reply.OpErrno = 0

	if (args.Flags & gfGetspecFlagServersList) != 0 {

		volinfo, err := volume.GetVolume(volfileID)
		if err != nil {
			log.WithError(err).WithField("volume", volfileID).Warn("failed to get volinfo from store")
			// Currently there's no easy way to distinguish between
			// a GETSPEC request from client vs request from daemon
			// such as self-heal as self-heal is also a client.
			err = nil
			goto Out
		}

		// We can return list of all peers too. That'll be correct
		// only if bricks too can connect to any glusterd2 to get its
		// volfile, which isn't true today.
		// For now, let's just return list of peers which has bricks
		// belonging to the volume being mounted.

		peers := volinfo.Peers()
		for _, p := range peers {
			for _, addr := range p.ClientAddresses {
				if !strings.HasPrefix(addr, "127.") && !strings.HasPrefix(addr, "localhost") {
					addrs = append(addrs, addr)
				}
			}
		}

		if len(addrs) > 0 {
			respDict = make(map[string]string)
			respDict["servers-list"] = strings.Join(addrs, " ")
			reply.Xdata, err = dict.Serialize(respDict)
			if err != nil {
				log.WithError(err).Error("failed to serialize dict")
			}
		}
	}

Out:
	if err != nil {
		reply.OpRet = -1
		reply.OpErrno = 0
	}

	return nil
}

// GfGetVolumeInfoReq is a request sent by glusterfs client. It contains a dict
// which contains information about the volume information requested by the
// client.
type GfGetVolumeInfoReq struct {
	Dict []byte
}

// GfGetVolumeInfoResp is response sent to glusterfs client in response to a
// GfGetVolumeInfoReq request. The dict shall contain actual information
// requested by the client.
type GfGetVolumeInfoResp struct {
	OpRet    int
	OpErrno  int
	OpErrstr string
	Dict     []byte
}

const gfGetVolumeUUID = 1

// ServerGetVolumeInfo returns requested information about the volume to the
// client.
func (p *GfHandshake) ServerGetVolumeInfo(args *GfGetVolumeInfoReq, reply *GfGetVolumeInfoResp) error {

	var (
		// pre-declared variables are required for goto statements
		err      error
		ok       bool
		volname  string
		flagsStr string
		flags    int
		volinfo  *volume.Volinfo
	)
	respDict := make(map[string]string)

	reqDict, err := dict.Unserialize(args.Dict)
	if err != nil {
		log.WithError(err).Error("dict unserialize failed")
		goto Out
	}

	flagsStr, ok = reqDict["flags"]
	if !ok {
		err = errors.New("flags key not found")
		goto Out
	}
	flags, err = strconv.Atoi(flagsStr)
	if err != nil {
		log.WithError(err).Error("failed to convert flags from string to int")
		goto Out
	}

	volname, ok = reqDict["volname"]
	if !ok {
		log.WithError(err).WithField("volume", volname).Error("volume name not found in request dict")
		reply.OpRet = -1
		reply.OpErrno = int(syscall.EINVAL)
		goto Out
	}

	if (flags & gfGetVolumeUUID) != 0 {
		volinfo, err = volume.GetVolume(volname)
		if err != nil {
			log.WithError(err).WithField("volume", volname).Error("volume not found in store")
			reply.OpErrno = int(syscall.EINVAL)
			goto Out
		}
		respDict["volume_id"] = volinfo.ID.String()
	}

	reply.Dict, err = dict.Serialize(respDict)
	if err != nil {
		log.WithError(err).Error("failed to serialize dict")
	}

Out:
	if err != nil {
		reply.OpRet = -1
		reply.OpErrstr = err.Error()
	}

	return nil
}

// GfServerEventNotifyReq is sent by the rebalance process before it terminates
// and contains the status information in a dict
type GfServerEventNotifyReq struct {
	Op   int
	Dict []byte
}

//GfServerEventNotifyResp contains the response to the GfServerEventNotifyReq
type GfServerEventNotifyResp struct {
	OpRet   int
	OpErrno int
	Dict    []byte
}

//ServerEventNotify processes the status information sent by the rebalance process
func (p *GfHandshake) ServerEventNotify(args *GfServerEventNotifyReq, reply *GfServerEventNotifyResp) error {

	var (
		// pre-declared variables are required for goto statements
		err error
	)

	switch args.Op {
	case gfEventNotifyDefragStatus:
		reqDict, err := dict.Unserialize(args.Dict)
		if err != nil {
			log.WithError(err).Error("dict unserialize failed")
			reply.OpRet = -1
			reply.OpErrno = int(syscall.EINVAL)
			goto Out
		}
		err = rebalance.HandleEventNotify(reqDict)
		if err != nil {
			reply.OpRet = -1
			reply.OpErrno = int(syscall.EINVAL)
			goto Out
		}

	default:
		log.WithError(err).Error("Unknown op received in event notify")
		reply.OpRet = -1
		reply.OpErrno = int(syscall.EINVAL)
		goto Out
	}

Out:
	if err != nil {
		reply.OpRet = -1
	}

	return nil

}
