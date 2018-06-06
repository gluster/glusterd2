package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
)

func volumeListHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	keys, keyFound := r.URL.Query()["key"]
	values, valueFound := r.URL.Query()["value"]
	filterParams := make(map[string]string)

	if keyFound {
		filterParams["key"] = keys[0]
	}
	if valueFound {
		filterParams["value"] = values[0]
	}
	volumes, err := volume.GetVolumes(filterParams)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
	}
	resp := createVolumeListResp(volumes)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeListResp(volumes []*volume.Volinfo) *api.VolumeListResp {
	var resp = make(api.VolumeListResp, len(volumes))

	for index, v := range volumes {
		resp[index] = *(createVolumeGetResp(v))
	}

	return &resp
}
