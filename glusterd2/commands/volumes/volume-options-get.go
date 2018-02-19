package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
)

func volumeOptionsGetHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	volname := mux.Vars(r)["volname"]
	v, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	resp := createVolumeOptionsGetResp(v)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeOptionsGetResp(v *volume.Volinfo) *api.VolumeOptionsGetResp {
	var resp api.VolumeOptionsGetResp

	for _, xl := range xlator.Xlators() {
		// TODO Once we have information on supported xlators
		// per volume type we can filter out these options. For
		// now return all options

		for _, opt := range xl.Options {
			var val string
			modified := false

			for _, k := range opt.Key {
				val = opt.DefaultValue

				if _, found := v.Options[k]; found {
					val = v.Options[k]
					modified = true
				}
				resp = append(resp, api.VolumeOptionGetResp{
					OptName:      xl.ID + "." + k,
					Value:        val,
					Modified:     modified,
					DefaultValue: opt.DefaultValue,
				})
			}

		}
	}
	return &resp
}
