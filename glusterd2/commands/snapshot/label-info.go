package snapshotcommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot/label"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gorilla/mux"
)

func labelInfoHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	labelname := mux.Vars(r)["labelname"]
	labelInfo, err := label.GetLabel(labelname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	resp := createLabelGetResp(labelInfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createLabelGetResp(info *label.Info) *api.LabelGetResp {
	return (*api.LabelGetResp)(label.CreateLabelInfoResp(info))
}
