package sunrpc

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/gluster/glusterd2/store"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/prashanthpai/sunrpc"
)

const (
	hndskProgNum     = 14398633 // GLUSTER_HNDSK_PROGRAM
	hndskProgVersion = 2        // GLUSTER_HNDSK_VERSION
)

const (
	gfHndskGetSpec = 2 // GF_HNDSK_GETSPEC
)

var volfilePrefix = store.GlusterPrefix + "volfiles/"

// GfHandshake is a type for GlusterFS Handshake RPC program
type GfHandshake genericProgram

func newGfHandshake() *GfHandshake {
	// rpc/rpc-lib/src/protocol-common.h
	return &GfHandshake{
		name:        "Gluster Handshake",
		progNum:     hndskProgNum,
		progVersion: hndskProgVersion,
		procedures: []sunrpc.Procedure{
			{
				sunrpc.ProcedureID{ProgramNumber: hndskProgNum, ProgramVersion: hndskProgVersion,
					ProcedureNumber: gfHndskGetSpec}, "ServerGetspec"},
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

// ServerGetspec returns the content of client volfile for the volume
// specified by the client
func (p *GfHandshake) ServerGetspec(args *GfGetspecReq, reply *GfGetspecRsp) error {
	var err error
	var fileContents []byte
	var volFilePath string

	xdata, err := DictUnserialize(args.Xdata)
	if err != nil {
		log.WithError(err).Error("ServerGetspec(): DictUnserialize() failed")
		goto Out
	}

	if _, ok := xdata["brick_name"]; ok {
		// brick volfile
		s := strings.Split(args.Key, ".")
		volName := s[0]
		volFilePath = path.Join(utils.GetVolumeDir(volName), fmt.Sprintf("%s.vol", args.Key))
		fileContents, err = ioutil.ReadFile(volFilePath)
		if err != nil {
			log.WithError(err).Error("ServerGetspec(): Could not read brick volfile")
			goto Out
		}
		log.Info(fileContents)
	} else {
		// client volfile
		resp, err := store.Store.Get(context.TODO(), volfilePrefix+args.Key)
		if err != nil {
			log.WithError(err).Error("ServerGetspec(): failed to retrive client volfile from store")
			goto Out
		}

		if resp.Count != 1 {
			log.WithField("volume", args.Key).Error("ServerGetspec(): client volfile not found in store")
			goto Out
		}

		fileContents = resp.Kvs[0].Value
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
