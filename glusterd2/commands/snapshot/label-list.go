package snapshotcommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot/label"
	"github.com/gluster/glusterd2/pkg/api"
)

func labelListHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	labelInfos, err := label.GetLabels()
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	resp := createLabelListResp(labelInfos)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createLabelListResp(infos []*label.Info) *api.LabelListResp {
	var resp = make(api.LabelListResp, len(infos))

	for index, v := range infos {
		resp[index] = *(createLabelGetResp(v))
	}

	return &resp
}
