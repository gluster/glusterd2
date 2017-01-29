package sunrpc

import (
	"net/rpc"
	"reflect"

	"github.com/prashanthpai/sunrpc"
)

// Procedure represents a procedure number, procedure name pair
type Procedure struct {
	Number uint32
	Name   string
}

// Program is an interface that every RPC program should implement
type Program interface {
	Name() string
	Number() uint32
	Version() uint32
	Procedures() []Procedure
}

// RPC program implementations can use this type for convenience
type genericProgram struct {
	name        string
	progNum     uint32
	progVersion uint32
	procedures  []Procedure
}

func registerProgram(server *rpc.Server, program Program, port int) error {

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
		err = sunrpc.RegisterProcedure(
			sunrpc.ProcedureID{
				program.Number(),
				program.Version(),
				procedure.Number,
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
