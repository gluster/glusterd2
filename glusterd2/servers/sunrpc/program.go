package sunrpc

import (
	"net/rpc"
	"reflect"
	"strings"

	"github.com/gluster/glusterd2/pkg/sunrpc"

	log "github.com/sirupsen/logrus"
)

// RPC program implementations inside this package can use this type for convenience
type genericProgram struct {
	name        string
	progNum     uint32
	progVersion uint32
	procedures  []sunrpc.Procedure
}

func registerProgram(server *rpc.Server, program sunrpc.Program, port int, tellrpcbind bool) error {
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
		log.WithFields(log.Fields{
			"procId":   procedure.ID,
			"procName": procedure.Name,
		}).Debug("registering sunrpc procedure")

		if !strings.HasPrefix(procedure.Name, typeName+".") {
			procedure.Name = typeName + "." + procedure.Name
		}
		err = sunrpc.RegisterProcedure(
			sunrpc.Procedure{
				ID:   procedure.ID,
				Name: procedure.Name,
			}, true)
		if err != nil {
			return err
		}
	}

	if tellrpcbind && port != 0 {
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
