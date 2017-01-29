package sunrpc

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
)

const (
	hndskProgNum     = 14398633
	hndskProgVersion = 2

	gfHndskGetSpec = 2
)

// GfHandshake is a type for GlusterFS Handshake RPC program
type GfHandshake genericProgram

func newGfHandshake() *GfHandshake {
	// rpc/rpc-lib/src/protocol-common.h
	return &GfHandshake{
		name:        "Gluster Handshake",
		progNum:     hndskProgNum,
		progVersion: hndskProgVersion,
		procedures: []Procedure{
			Procedure{gfHndskGetSpec, "ServerGetspec"}, // GF_HNDSK_GETSPEC
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
func (p *GfHandshake) Procedures() []Procedure {
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
