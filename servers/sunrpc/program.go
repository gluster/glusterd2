package sunrpc

import (
	"net/rpc"
	"reflect"

	"github.com/gluster/glusterd2/servers/sunrpc/program"
	log "github.com/Sirupsen/logrus"
	"github.com/prashanthpai/sunrpc"
)

// RPC program implementations can use this type for convenience
type genericProgram struct {
	name        string
	progNum     uint32
	progVersion uint32
	procedures  []program.Procedure
}

func registerProgram(server *rpc.Server, program program.Program, port int) error {
	logger := log.WithFields(log.Fields{
		"program": program.Name(),
		"prognum": program.Number(),
		"progver": program.Version(),
	})

	logger.Debug("registering sunrpc program")

	// NOTE: This will throw some benign log messages complaining about
	// signatures of methods in Program interface. rpc.Server.Register()
	// expects all methods of program to be of the kind:
	//         func (t *T) MethodName(argType T1, replyType *T2) error
	// These log entries (INFO) can be ignored.
	err := server.Register(program)
	if err != nil {
		return err
	}

	// Create procedure number to procedure name mappings for sunrpc codec
	typeName := reflect.Indirect(reflect.ValueOf(program)).Type().Name()
	for _, procedure := range program.Procedures() {
		logger.WithFields(log.Fields{
			"proc":    procedure.Name,
			"procnum": procedure.Number,
		}).Debug("registering sunrpc procedure")

		err = sunrpc.RegisterProcedure(
			sunrpc.ProcedureID{
				ProgramNumber:   program.Number(),
				ProgramVersion:  program.Version(),
				ProcedureNumber: procedure.Number,
			}, typeName+"."+procedure.Name)
		if err != nil {
			return err
		}
	}

	if port != 0 {
		_, err = sunrpc.PmapUnset(program.Number(), program.Version())
		if err != nil {
			return err
		}

		_, err = sunrpc.PmapSet(program.Number(), program.Version(), sunrpc.IPProtoTCP, uint32(port))
		if err != nil {
			return err
		}
	}

	return nil
}
