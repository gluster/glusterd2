package volumecommands

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/glusterd2/xlator/options"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/gorilla/mux"
)

func volumeOptionsGetHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	volname := mux.Vars(r)["volname"]

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	optname, found := mux.Vars(r)["optname"]

	if found {
		opt, err := xlator.FindOption(optname)
		if err != nil {
			if _, ok := err.(xlator.OptionNotFoundError); ok {
				restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
				return
			}
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
			return
		}
		resp := createVolumeOptionGetResp(volinfo, opt, optname)
		restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
	} else {
		resp := createVolumeOptionsGetResp(volinfo)
		restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
	}
}

func createVolumeOptionGetResp(volinfo *volume.Volinfo, opt *options.Option, optname string) *api.VolumeOptionGetResp {
	var (
		resp     api.VolumeOptionGetResp
		modified bool
		optValue string
	)
	optValue = opt.DefaultValue

	if value, ok := volinfo.Options[optname]; ok {
		modified = true
		optValue = value
	}

	resp = api.VolumeOptionGetResp{
		OptName:      optname,
		Value:        optValue,
		Modified:     modified,
		DefaultValue: opt.DefaultValue,
		OptionLevel:  opt.Level.String(),
	}

	return &resp
}

func createVolumeOptionsGetResp(volinfo *volume.Volinfo) *api.VolumeOptionsGetResp {
	var resp api.VolumeOptionsGetResp

	for _, xl := range xlator.Xlators() {
		// TODO Once we have information on supported xlators
		// per volume type we can filter out these options. For
		// now return all options
		for _, opt := range xl.Options {
			for _, k := range opt.Key {
				optName := xl.ID + "." + k
				volOptRest := createVolumeOptionGetResp(volinfo, opt, optName)
				resp = append(resp, *volOptRest)
			}
		}
	}
	return &resp
}
