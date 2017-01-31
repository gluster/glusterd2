package sunrpc

// GfCommonRsp is a generic RPC response type
type GfCommonRsp struct {
	OpRet   int
	OpErrno int
	Xdata   []byte
}
