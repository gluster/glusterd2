package brick

import (
	"github.com/prashanthpai/sunrpc"
)

const (
	brickOpsProgNum     = 4867634 // GD_BRICK_PROGRAM
	brickOpsProgVersion = 2       // GD_BRICK_VERSION
	brickOp             = 1       // GLUSTERD_BRICK_OP
)

// These are op numbers and not separate RPC procedures
const (
	OpBrickNull         = iota // GLUSTERD_BRICK_NULL
	OpBrickTerminate           // GLUSTERD_BRICK_TERMINATE
	OpBrickXlatorInfo          // GLUSTERD_BRICK_XLATOR_INFO
	OpBrickXlatorOp            // GLUSTERD_BRICK_XLATOR_OP
	OpBrickStatus              // GLUSTERD_BRICK_STATUS
	OpBrickOp                  // GLUSTERD_BRICK_OP
	OpBrickXlatorDefrag        // GLUSTERD_BRICK_XLATOR_DEFRAG
	OpNodeProfile              // GLUSTERD_NODE_PROFILE
	OpNodeStatus               // GLUSTERD_NODE_STATUS
	OpVolumeBarrierOp          // GLUSTERD_VOLUME_BARRIER_OP
	OpBrickBarrier             // GLUSTERD_BRICK_BARRIER
	OpNodeBitrot               // GLUSTERD_NODE_BITROT
	OpBrickAttach              // GLUSTERD_BRICK_ATTACH
)

func init() {
	pID := sunrpc.ProcedureID{
		ProgramNumber:   brickOpsProgNum,
		ProgramVersion:  brickOpsProgVersion,
		ProcedureNumber: brickOp,
	}
	_ = sunrpc.RegisterProcedure(sunrpc.Procedure{
		ID:   pID,
		Name: "BrickOp",
	}, false)
}
