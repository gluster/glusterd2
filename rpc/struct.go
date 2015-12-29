package rpc

type RPCResponse struct {
	OpRet   int
	OpError string
	RspMap  map[string]string
}

type Connection int
