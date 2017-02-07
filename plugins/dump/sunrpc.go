package dumpplugin

import (
	"github.com/gluster/glusterd2/servers/sunrpc/program"
	"github.com/gluster/glusterd2/servers/sunrpc/common"
)

const (
	dumpProgName    = "GF-DUMP"
	dumpProgNum     = 123451501
	dumpProgVersion = 1
)

const (
	_ = iota
	gfDumpDump
	gfDumpPing
)

// GfDump is a type for GlusterFS Dump RPC program
// rpc/rpc-lib/src/xdr-common.h
type GfDump struct{}

// Name returns the name of the RPC program
func (p *GfDump) Name() string {
	return dumpProgName
}

// Number returns the RPC Program number
func (p *GfDump) Number() uint32 {
	return dumpProgNum
}

// Version returns the RPC program version number
func (p *GfDump) Version() uint32 {
	return dumpProgVersion
}

// Procedures returns a list of procedures provided by the RPC program
func (p *GfDump) Procedures() []program.Procedure {
	return []program.Procedure{
		program.Procedure{gfDumpDump, "Dump"}, // GF_DUMP_DUMP
		program.Procedure{gfDumpPing, "Ping"}, // GF_DUMP_PING
	}
}

// GfDumpReq is request sent by the client
type GfDumpReq struct {
	GfsID uint64
}

// GfProcDetail contains details for individual RPC program
type GfProcDetail struct {
	ProgName string
	ProgNum  uint64
	ProgVer  uint64
	Next     *GfProcDetail `xdr:"optional"`
}

// GfDumpRsp is response sent by server. It contains a list of GfProcDetail
type GfDumpRsp struct {
	GfsID   uint64
	OpRet   int
	OpErrno int
	Prog    *GfProcDetail `xdr:"optional"`
}

// Dump will return a list of all available RPC programs
func (p *GfDump) Dump(args *GfDumpReq, reply *GfDumpRsp) error {

	// TODO: I don't like doing this in Go. Should abstract it.
	var list *GfProcDetail
	var trav *GfProcDetail

	for _, p := range program.ProgramsList {
		tmp := &GfProcDetail{
			ProgName: p.Name(),
			ProgNum:  uint64(p.Number()),
			ProgVer:  uint64(p.Version()),
		}
		if list == nil {
			list = tmp
			trav = list
		} else {
			trav.Next = tmp
			trav = trav.Next
		}
	}
	reply.Prog = list

	return nil
}

// Ping is for availability check
func (p *GfDump) Ping(_ *struct{}, reply *common.GfCommonRsp) error {

	return nil
}
