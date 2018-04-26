package pmap

import (
	"github.com/gluster/glusterd2/pkg/sunrpc"
)

const (
	portmapProgNum     = 34123456 // GLUSTER_PMAP_PROGRAM
	portmapProgVersion = 1        // GLUSTER_PMAP_VERSION
)

const (
	gfPmapNull        = iota
	gfPmapPortByBrick // GF_PMAP_PORTBYBRICK
	gfPmapBrickByPort // GF_PMAP_BRICKBYPORT, Not Implemented
	gfPmapSignUp      // Don't use
	gfPmapSignIn      // GF_PMAP_SIGNIN
	gfPmapSignOut     // GF_PMAP_SIGNOUT
)

// GfPortmap is a type for GlusterFS Portmap RPC program
type GfPortmap struct {
	name        string
	progNum     uint32
	progVersion uint32
	procedures  []sunrpc.Procedure
}

// NewGfPortmap returns a new instance of GfPortmap type
func NewGfPortmap() *GfPortmap {
	// rpc/rpc-lib/src/protocol-common.h
	return &GfPortmap{
		name:        "Gluster Portmap",
		progNum:     portmapProgNum,
		progVersion: portmapProgVersion,
		procedures: []sunrpc.Procedure{
			{
				ID: sunrpc.ProcedureID{ProgramNumber: portmapProgNum, ProgramVersion: portmapProgVersion,
					ProcedureNumber: gfPmapPortByBrick}, Name: "PortByBrick"},
			{
				ID: sunrpc.ProcedureID{ProgramNumber: portmapProgNum, ProgramVersion: portmapProgVersion,
					ProcedureNumber: gfPmapSignIn}, Name: "SignIn"},
			{
				ID: sunrpc.ProcedureID{ProgramNumber: portmapProgNum, ProgramVersion: portmapProgVersion,
					ProcedureNumber: gfPmapSignOut}, Name: "SignOut"},
		},
	}
}

// Name returns the name of the RPC program
func (p *GfPortmap) Name() string {
	return p.name
}

// Number returns the RPC Program number
func (p *GfPortmap) Number() uint32 {
	return p.progNum
}

// Version returns the RPC program version number
func (p *GfPortmap) Version() uint32 {
	return p.progVersion
}

// Procedures returns a list of procedures provided by the RPC program
func (p *GfPortmap) Procedures() []sunrpc.Procedure {
	return p.procedures
}

// PortByBrickReq is sent by the glusterfs client
type PortByBrickReq struct {
	Brick string
}

// PortByBrickRsp is sent to glusterfs client and contains the port
// for the brick requested
type PortByBrickRsp struct {
	OpRet   int
	OpErrno int
	Status  int
	Port    int
}

// PortByBrick will return port number for the brick specified
func (p *GfPortmap) PortByBrick(args *PortByBrickReq, reply *PortByBrickRsp) error {

	port := RegistrySearch(args.Brick, GfPmapPortBrickserver)
	if port <= 0 {
		reply.OpRet = -1
	} else {
		reply.Port = port
	}

	return nil
}

// SignInReq is the request received
type SignInReq struct {
	Brick string
	Port  int
}

// SignInRsp is response sent to a SignInReq request
type SignInRsp struct {
	OpRet   int
	OpErrno int
}

// SignIn stores the brick and port mapping in registry
func (p *GfPortmap) SignIn(args *SignInReq, reply *SignInRsp) error {

	// FIXME: Xprt (net.Conn instance) isn't available here yet.
	// Passing nil for now.
	registryBind(args.Port, args.Brick, GfPmapPortBrickserver, nil)

	return nil
}

// SignOutReq is the request received
type SignOutReq struct {
	Brick    string
	Port     int
	RdmaPort int
}

// SignOutRsp is response sent to a SignOutReq request
type SignOutRsp struct {
	OpRet   int
	OpErrno int
}

// SignOut removes the brick and port mapping in registry
func (p *GfPortmap) SignOut(args *SignOutReq, reply *SignOutRsp) error {

	// FIXME: Xprt (net.Conn instance) isn't available here yet.
	// Passing nil for now.
	registryRemove(args.Port, args.Brick, GfPmapPortBrickserver, nil)

	return nil
}
