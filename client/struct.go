package client

import (
	"encoding/json"
	"net/http"
)

// GenericJSONResponse defines the generic response type to be sent back to
// client
type GenericJSONResponse struct {
	OpRet    int
	OpErrno  int
	OpErrstr string
	Data     interface{}
}

// SendResponse Warapper function to send Generic JSON response back to the rest client
func SendResponse(w http.ResponseWriter, opRet int, opErrno int, opErrStr string, status int, v interface{}) {
	var rsp GenericJSONResponse

	rsp.OpRet = opRet
	rsp.OpErrno = opErrno
	rsp.OpErrstr = opErrStr
	w.WriteHeader(status)
	rsp.Data = v

	if e := json.NewEncoder(w).Encode(rsp); e != nil {
		panic(e)
	}
	return
}
