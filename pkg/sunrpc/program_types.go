package sunrpc

// Program is an interface that every RPC program can implement and
// use internally for convenience during procedure registration
type Program interface {
	Name() string
	Number() uint32
	Version() uint32
	Procedures() []Procedure
}
