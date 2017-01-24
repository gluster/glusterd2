package sunrpc

import (
	"fmt"
	"io/ioutil"
	"net/rpc"
	"path"

	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/prashanthpai/sunrpc"
)

// TODO: The enum names here are copied from gluster code for simplicity.
// golint checks will fail.

const (
	GLUSTER_HNDSK_PROGRAM uint32 = 14398633
	GLUSTER_HNDSK_VERSION uint32 = 2
)

const (
	// rpc/rpc-lib/src/protocol-common.h
	GF_HNDSK_NULL uint32 = iota
	GF_HNDSK_SETVOLUME
	GF_HNDSK_GETSPEC // 2
	GF_HNDSK_PING
	GF_HNDSK_SET_LK_VER
	GF_HNDSK_EVENT_NOTIFY
	GF_HNDSK_GET_VOLUME_INFO
	GF_HNDSK_GET_SNAPSHOT_INFO
	GF_HNDSK_MAXVALUE
)

func registerHandshakeProgram(server *rpc.Server, port int) error {

	err := server.Register(new(GfHandshake))
	if err != nil {
		return err
	}

	// TODO: As the number of procedures to be registered grows, we need
	// to make this boilerplate code less verbose by following some
	// naming conventions for the enums or procedures themselves.
	err = sunrpc.RegisterProcedure(
		sunrpc.ProcedureID{
			GLUSTER_HNDSK_PROGRAM,
			GLUSTER_HNDSK_VERSION,
			GF_HNDSK_GETSPEC,
		},
		"GfHandshake.ServerGetspec")
	if err != nil {
		return err
	}

	if port != 0 {
		_, err = sunrpc.PmapUnset(GLUSTER_HNDSK_PROGRAM, GLUSTER_HNDSK_VERSION)
		if err != nil {
			return err
		}

		_, err = sunrpc.PmapSet(
			GLUSTER_HNDSK_PROGRAM, GLUSTER_HNDSK_VERSION,
			sunrpc.IPProtoTCP, uint32(port))
		if err != nil {
			return err
		}
	}

	return nil
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

// GfHandshake is a placeholder type for all functions of GlusterFS Handshake
// RPC program
type GfHandshake int32

// ServerGetspec returns the content of client volfile for the volume
// specified by the client
func (t *GfHandshake) ServerGetspec(args *GfGetspecReq, reply *GfGetspecRsp) error {
	var err error
	var fileContents []byte
	var volFilePath string

	_, err = DictUnserialize(args.Xdata)
	if err != nil {
		log.WithError(err).Error("ServerGetspec(): DictUnserialize() failed")
		goto Out
	}

	volFilePath = path.Join(utils.GetVolumeDir(args.Key), fmt.Sprintf("trusted-%s.tcp-fuse.vol", args.Key))
	fileContents, err = ioutil.ReadFile(volFilePath)
	if err != nil {
		log.WithError(err).Error("ServerGetspec(): Could not read client volfile")
		goto Out
	}
	reply.Spec = string(fileContents)
	reply.OpRet = len(reply.Spec)
	reply.OpErrno = 0

Out:
	if err != nil {
		reply.OpRet = -1
		reply.OpErrno = 0
	}

	return nil
}
