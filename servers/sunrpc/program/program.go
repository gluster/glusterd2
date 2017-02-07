package program

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

var ProgramsList []Program
