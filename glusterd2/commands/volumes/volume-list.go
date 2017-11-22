package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
)

func volumeListHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	volumes, err := volume.GetVolumes()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err.Error(), api.ErrCodeDefault)
	}

	resp := createVolumeListResp(volumes)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeListResp(volumes []*volume.Volinfo) *api.VolumeListResp {
	var resp api.VolumeListResp

	for _, v := range volumes {
		resp = append(resp, *(createVolumeGetResp(v)))
	}

	return &resp
}
