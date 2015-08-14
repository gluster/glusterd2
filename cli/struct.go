package cli

import (
	"encoding/json"
	"net/http"
)

type GenericJsonResponse struct {
	OpRet    int
	OpErrno  int
	OpErrstr string
	Data     interface{}
}

func SendResponse(w http.ResponseWriter, opRet int, opErrno int, opErrstr string, status int, v interface{}) {
	var rsp GenericJsonResponse

	rsp.OpRet = opRet
	rsp.OpErrno = opErrno
	rsp.OpErrstr = opErrstr
	w.WriteHeader(status)
	rsp.Data = v

	if e := json.NewEncoder(w).Encode(rsp); e != nil {
		panic(e)
	}
	return
}
