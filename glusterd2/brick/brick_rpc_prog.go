package brick

import (
	"github.com/gluster/glusterd2/pkg/sunrpc"
)

const (
	brickOpsProgNum     = 4867634 // GD_BRICK_PROGRAM
	brickOpsProgVersion = 2       // GD_BRICK_VERSION
)

// TypeBrickOp is a brick operation
//go:generate stringer -type=TypeBrickOp
type TypeBrickOp uint32

// RPC procedures
const (
	OpBrickNull         TypeBrickOp = iota // GLUSTERD_BRICK_NULL
	OpBrickTerminate                       // GLUSTERD_BRICK_TERMINATE
	OpBrickXlatorInfo                      // GLUSTERD_BRICK_XLATOR_INFO
	OpBrickXlatorOp                        // GLUSTERD_BRICK_XLATOR_OP
	OpBrickStatus                          // GLUSTERD_BRICK_STATUS
	OpBrickOp                              // GLUSTERD_BRICK_OP
	OpBrickXlatorDefrag                    // GLUSTERD_BRICK_XLATOR_DEFRAG
	OpNodeProfile                          // GLUSTERD_NODE_PROFILE
	OpNodeStatus                           // GLUSTERD_NODE_STATUS
	OpVolumeBarrierOp                      // GLUSTERD_VOLUME_BARRIER_OP
	OpBrickBarrier                         // GLUSTERD_BRICK_BARRIER
	OpNodeBitrot                           // GLUSTERD_NODE_BITROT
	OpBrickAttach                          // GLUSTERD_BRICK_ATTACH
	OpDumpMetrics                          // GLUSTERD_DUMP_METRICS
	OpMaxValue                             // GLUSTERD_BRICK_MAXVALUE
)

// GfBrickOpReq is the request sent to the brick process
type GfBrickOpReq struct {
	Name  string
	Op    int
	Input []byte
}

// GfBrickOpRsp is the response sent by brick to a BrickOpReq request
type GfBrickOpRsp struct {
	OpRet    int
	OpErrno  int
	Output   []byte
	OpErrstr string
}

func init() {
	var p sunrpc.Procedure
	for i := 0; i < int(OpMaxValue); i++ {
		p.ID = sunrpc.ProcedureID{
			ProgramNumber:   brickOpsProgNum,
			ProgramVersion:  brickOpsProgVersion,
			ProcedureNumber: uint32(i),
		}
		p.Name = "Brick." + TypeBrickOp(i).String()
		_ = sunrpc.RegisterProcedure(p, false)
	}
}
