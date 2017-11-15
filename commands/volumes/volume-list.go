package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/pkg/api"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/volume"
)

func volumeListHandler(w http.ResponseWriter, r *http.Request) {

	volumes, err := volume.GetVolumes()
	if err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error())
	}

	resp := createVolumeListResp(volumes)
	restutils.SendHTTPResponse(w, http.StatusOK, resp)
}

func createVolumeListResp(volumes []*volume.Volinfo) *api.VolumeListResp {
	var resp api.VolumeListResp

	for _, v := range volumes {
		resp = append(resp, *(createVolumeGetResp(v)))
	}

	return &resp
}
