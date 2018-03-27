package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/gorilla/mux"
)

func volumeOptionGetHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	volname := mux.Vars(r)["volname"]
	optname := mux.Vars(r)["optname"]

	v, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
		return
	}

	var resp api.VolumeOptionGetResp
	for _, xl := range xlator.Xlators() {
		// TODO Once we have information on supported xlators
		// per volume type we can filter out these options. For
		// now return all options

		opt, err := xlator.FindOption(xl.ID + "." + optname)
		if err != nil {
			continue
		}

		modified, found := false, false
		val := opt.DefaultValue

		for _, k := range opt.Key {
			if val, found = v.Options[xl.ID+"."+k]; found {
				modified = found
			}
			resp = api.VolumeOptionGetResp{
				OptName:      xl.ID + "." + k,
				Value:        val,
				Modified:     modified,
				DefaultValue: opt.DefaultValue,
				OptLevel:     api.OptionLevel(opt.Level),
			}
			break
		}
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}
