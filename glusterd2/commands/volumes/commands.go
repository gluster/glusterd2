// Package volumecommands implements the volume management commands
package volumecommands

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"
)

// Command is a holding struct used to implement the GlusterD Command interface
type Command struct {
}

// Routes returns command routes. Required for the Command interface.
func (c *Command) Routes() route.Routes {
	return route.Routes{
		route.Route{
			Name:         "VolumeCreate",
			Method:       "POST",
			Pattern:      "/volumes",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.VolCreateReq)(nil)),
			ResponseType: utils.GetTypeString((*api.VolumeCreateResp)(nil)),
			HandlerFunc:  volumeCreateHandler},
		route.Route{
			Name:         "VolumeExpand",
			Method:       "POST",
			Pattern:      "/volumes/{volname}/expand",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.VolExpandReq)(nil)),
			ResponseType: utils.GetTypeString((*api.VolumeExpandResp)(nil)),
			HandlerFunc:  volumeExpandHandler},
		route.Route{
			Name:         "VolumeOptionGet",
			Method:       "GET",
			Pattern:      "/volumes/{volname}/options/{optname}",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.VolumeOptionGetResp)(nil)),
			HandlerFunc:  volumeOptionsGetHandler},
		route.Route{
			Name:         "VolumeOptionsGet",
			Method:       "GET",
			Pattern:      "/volumes/{volname}/options",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.VolumeOptionsGetResp)(nil)),
			HandlerFunc:  volumeOptionsGetHandler},
		route.Route{
			Name:         "VolumeOptions",
			Method:       "POST",
			Pattern:      "/volumes/{volname}/options",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.VolOptionReq)(nil)),
			ResponseType: utils.GetTypeString((*api.VolumeOptionResp)(nil)),
			HandlerFunc:  volumeOptionsHandler},
		route.Route{
			Name:         "VolumeReset",
			Method:       "DELETE", // Do DELETE requests have a body? Should this be query param ?
			Pattern:      "/volumes/{volname}/options",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.VolOptionResetReq)(nil)),
			ResponseType: utils.GetTypeString((*api.VolumeOptionResp)(nil)),
			HandlerFunc:  volumeResetHandler},
		route.Route{
			Name:         "OptionGroupList",
			Method:       "GET",
			Pattern:      "/volumes/options-group",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.OptionGroupListResp)(nil)),
			HandlerFunc:  optionGroupListHandler},
		route.Route{
			Name:        "OptionGroupCreate",
			Method:      "POST",
			Pattern:     "/volumes/options-group",
			Version:     1,
			RequestType: utils.GetTypeString((*api.OptionGroupReq)(nil)),
			HandlerFunc: optionGroupCreateHandler},
		route.Route{
			Name:        "OptionGroupDelete",
			Method:      "DELETE",
			Pattern:     "/volumes/options-group/{groupname}",
			Version:     1,
			HandlerFunc: optionGroupDeleteHandler},
		route.Route{
			Name:        "VolumeDelete",
			Method:      "DELETE",
			Pattern:     "/volumes/{volname}",
			Version:     1,
			HandlerFunc: volumeDeleteHandler},
		route.Route{
			Name:         "VolumeInfo",
			Method:       "GET",
			Pattern:      "/volumes/{volname}",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.VolumeGetResp)(nil)),
			HandlerFunc:  volumeInfoHandler},
		route.Route{
			Name:         "VolumeBricksStatus",
			Method:       "GET",
			Pattern:      "/volumes/{volname}/bricks",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.BricksStatusResp)(nil)),
			HandlerFunc:  volumeBricksStatusHandler},
		route.Route{
			Name:         "VolumeStatus",
			Method:       "GET",
			Pattern:      "/volumes/{volname}/status",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.VolumeStatusResp)(nil)),
			HandlerFunc:  volumeStatusHandler},
		route.Route{
			Name:         "VolumeList",
			Method:       "GET",
			Pattern:      "/volumes",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.VolumeListResp)(nil)),
			HandlerFunc:  volumeListHandler},
		route.Route{
			Name:         "VolumeStart",
			Method:       "POST",
			Pattern:      "/volumes/{volname}/start",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.VolumeStartReq)(nil)),
			ResponseType: utils.GetTypeString((*api.VolumeStartResp)(nil)),
			HandlerFunc:  volumeStartHandler},
		route.Route{
			Name:         "VolumeStop",
			Method:       "POST",
			Pattern:      "/volumes/{volname}/stop",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.VolumeStopResp)(nil)),
			HandlerFunc:  volumeStopHandler},
		route.Route{
			Name:        "Statedump",
			Method:      "POST",
			Pattern:     "/volumes/{volname}/statedump",
			Version:     1,
			RequestType: utils.GetTypeString((*api.VolStatedumpReq)(nil)),
			HandlerFunc: volumeStatedumpHandler},
		route.Route{
			Name:        "VolfilesGenerate",
			Method:      "POST",
			Pattern:     "/volfiles",
			Version:     1,
			HandlerFunc: volfilesGenerateHandler},
		route.Route{
			Name:        "VolfilesGet",
			Method:      "GET",
			Pattern:     "/volfiles",
			Version:     1,
			HandlerFunc: volfilesListHandler},
		route.Route{
			Name:        "VolfilesGet",
			Method:      "GET",
			Pattern:     "/volfiles/{volfileid:.*}",
			Version:     1,
			HandlerFunc: volfileGetHandler},
		route.Route{
			Name:         "EditVolume",
			Method:       "POST",
			Pattern:      "/volumes/{volname}/edit",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.VolEditReq)(nil)),
			ResponseType: utils.GetTypeString((*api.VolumeEditResp)(nil)),
			HandlerFunc:  volumeEditHandler},
	}
}

// RegisterStepFuncs implements a required function for the Command interface
func (c *Command) RegisterStepFuncs() {
	registerVolCreateStepFuncs()
	registerVolDeleteStepFuncs()
	registerVolStartStepFuncs()
	registerVolStopStepFuncs()
	registerBricksStatusStepFuncs()
	registerVolExpandStepFuncs()
	registerVolOptionStepFuncs()
	registerVolOptionResetStepFuncs()
	registerVolStatedumpFuncs()
}
