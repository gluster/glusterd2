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

// FormResponse Warapper function to construct Generic JSON response to be sent
// back to the rest client
func FormResponse(opRet int, opErrno int, opErrstr string, v interface{}) *GenericJSONResponse {
	rsp := new(GenericJSONResponse)

	rsp.OpRet = opRet
	rsp.OpErrno = opErrno
	rsp.OpErrstr = opErrstr
	rsp.Data = v

	return rsp
}

// SendResponse Warapper function to send Generic JSON response back to the rest client
func SendResponse(w http.ResponseWriter, status int, rsp *GenericJSONResponse) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if e := json.NewEncoder(w).Encode(rsp); e != nil {
		panic(e)
	}
	return
}
