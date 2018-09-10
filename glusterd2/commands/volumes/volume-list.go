package volumecommands

import (
	"context"
	"net/http"
	"strconv"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"

	"go.opencensus.io/trace"
)

func volumeListHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, span := trace.StartSpan(ctx, "/volumeListHandler")
	defer span.End()

	keys, keyFound := r.URL.Query()["key"]
	values, valueFound := r.URL.Query()["value"]
	filterParams := make(map[string]string)

	if keyFound {
		filterParams["key"] = keys[0]
	}
	if valueFound {
		filterParams["value"] = values[0]
	}
	volumes, err := volume.GetVolumes(ctx, filterParams)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	// Add the count of volumes being listed as an attribute in the span
	span.AddAttributes(
		trace.StringAttribute("numVols", strconv.Itoa(len(volumes))),
	)

	resp := createVolumeListResp(ctx, volumes)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeListResp(ctx context.Context, volumes []*volume.Volinfo) *api.VolumeListResp {
	ctx, span := trace.StartSpan(ctx, "createVolumeListResp")
	defer span.End()

	var resp = make(api.VolumeListResp, len(volumes))

	for index, v := range volumes {
		resp[index] = *(createVolumeGetResp(v))
	}

	return &resp
}
