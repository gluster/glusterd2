package brick

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
