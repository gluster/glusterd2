package sunrpc

const (
	portmapProgNum     = 34123456
	portmapProgVersion = 1

	gfPmapPortByBrick = 1
)

// GfPortmap is a type for GlusterFS Portmap RPC program
type GfPortmap genericProgram

func newGfPortmap() *GfPortmap {
	// rpc/rpc-lib/src/protocol-common.h
	return &GfPortmap{
		name:        "Gluster Portmap",
		progNum:     portmapProgNum,
		progVersion: portmapProgVersion,
		procedures: []Procedure{
			Procedure{gfPmapPortByBrick, "PortByBrick"}, // GF_PMAP_PORTBYBRICK
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
func (p *GfPortmap) Procedures() []Procedure {
	return p.procedures
}

// PmapPortByBrickReq is sent by the glusterfs client
type PmapPortByBrickReq struct {
	Brick string
}

// PmapPortByBrickRsp is sent to glusterfs client and contains the port
// for the brick requested
type PmapPortByBrickRsp struct {
	OpRet   int
	OpErrno int
	Status  int
	Port    int
}

// PortByBrick will return port number for the brick specified
func (p *GfPortmap) PortByBrick(args *PmapPortByBrickReq, reply *PmapPortByBrickRsp) error {
	// TODO: Do the real thing. Glusterd2 as of now, doesn't store brick
	// port information in brickinfo. So can't return the ports. The
	// following code just demonstrates that when glusterd2 does have
	// that information, this will just work.

	switch {
	case args.Brick == "/export/brick1/data":
		reply.Port = 49152
	case args.Brick == "/export/brick2/data":
		reply.Port = 49153
	case args.Brick == "/export/brick3/data":
		reply.Port = 49154
	case args.Brick == "/export/brick4/data":
		reply.Port = 49155
	}

	return nil
}
