package snapshotcommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"
)

// Command is a structure which implements GlusterD Command interface
type Command struct {
}

// Routes returns list of REST API routes to register with Glusterd
func (c *Command) Routes() route.Routes {
	return route.Routes{
		route.Route{
			Name:         "SnapshotCreate",
			Method:       "POST",
			Pattern:      "/snapshots",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.SnapCreateReq)(nil)),
			ResponseType: utils.GetTypeString((*api.SnapCreateResp)(nil)),
			HandlerFunc:  snapshotCreateHandler},
		route.Route{
			Name:         "SnapshotActivate",
			Method:       "POST",
			Pattern:      "/snapshots/{snapname}/activate",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.SnapActivateReq)(nil)),
			ResponseType: utils.GetTypeString((*api.SnapshotActivateResp)(nil)),
			HandlerFunc:  snapshotActivateHandler},
		route.Route{
			Name:         "SnapshotDeactivate",
			Method:       "POST",
			Pattern:      "/snapshots/{snapname}/deactivate",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.SnapshotDeactivateResp)(nil)),
			HandlerFunc:  snapshotDeactivateHandler},
		route.Route{
			Name:         "SnapshotClone",
			Method:       "POST",
			Pattern:      "/snapshots/{snapname}/clone",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.SnapCloneReq)(nil)),
			ResponseType: utils.GetTypeString((*api.SnapshotCloneResp)(nil)),
			HandlerFunc:  snapshotCloneHandler},
		route.Route{
			Name:        "SnapshotRestore",
			Method:      "POST",
			Pattern:     "/snapshots/{snapname}/restore",
			Version:     1,
			HandlerFunc: snapshotRestoreHandler},
		route.Route{
			Name:         "SnapshotInfo",
			Method:       "GET",
			Pattern:      "/snapshots/{snapname}",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.SnapGetResp)(nil)),
			HandlerFunc:  snapshotInfoHandler},
		route.Route{
			Name:         "SnapshotListAll",
			Method:       "GET",
			Pattern:      "/snapshots",
			Version:      1,
			HandlerFunc:  snapshotListHandler,
			ResponseType: utils.GetTypeString((*api.SnapListResp)(nil))},
		route.Route{
			Name:         "SnapshotStatus",
			Method:       "GET",
			Pattern:      "/snapshots/{snapname}/status",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.SnapStatusResp)(nil)),
			HandlerFunc:  snapshotStatusHandler},
		route.Route{
			Name:        "SnapshotDelete",
			Method:      "DELETE",
			Pattern:     "/snapshots/{snapname}",
			Version:     1,
			HandlerFunc: snapshotDeleteHandler},
		route.Route{
			Name:        "SnapshotConfigGet",
			Method:      "GET",
			Pattern:     "/snapshots/config",
			Version:     1,
			HandlerFunc: snapshotConfigGetHandler},
		route.Route{
			Name:        "SnapshotConfigSet",
			Method:      "POST",
			Pattern:     "/snapshots/config",
			Version:     1,
			HandlerFunc: snapshotConfigSetHandler},
		route.Route{
			Name:        "SnapshotConfigReset",
			Method:      "DELETE",
			Pattern:     "/snapshots/config",
			Version:     1,
			HandlerFunc: snapshotConfigResetHandler},
	}
}

func snapshotConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusNotImplemented, "Snapshot Config Get")
}

func snapshotConfigSetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusNotImplemented, "Snapshot Config Set")
}

func snapshotConfigResetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusNotImplemented, "Snapshot Config Reset")
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (c *Command) RegisterStepFuncs() {
	registerSnapCreateStepFuncs()
	registerSnapActivateStepFuncs()
	registerSnapDeactivateStepFuncs()
	registerSnapDeleteStepFuncs()
	registerSnapshotStatusStepFuncs()
	registerSnapRestoreStepFuncs()
	registerSnapCloneStepFuncs()
	return
}
