package sunrpc

import (
	"net"
)

// Conn is an interface that RPC programs can implement that has setter and
// getter methods that set and return the underlying net.Conn connection
// object. This is a hack/workaround because net/rpc doesn't provide RPC
// programs access to net.Conn or context.Context.
type Conn interface {
	GetConn() net.Conn
	SetConn(net.Conn)
}
