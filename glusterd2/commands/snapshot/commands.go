package snapshotcommands

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
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
			Name:         "LabelCreate",
			Method:       "POST",
			Pattern:      "/snapshots/labels/create",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.LabelCreateReq)(nil)),
			ResponseType: utils.GetTypeString((*api.LabelCreateResp)(nil)),
			HandlerFunc:  labelCreateHandler},
		route.Route{
			Name:         "LabelInfo",
			Method:       "GET",
			Pattern:      "/snapshots/labels/{labelname}",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.LabelGetResp)(nil)),
			HandlerFunc:  labelInfoHandler},
		route.Route{
			Name:         "LabelListAll",
			Method:       "GET",
			Pattern:      "/snapshots/labels/list/all",
			Version:      1,
			ResponseType: utils.GetTypeString((*api.LabelListResp)(nil)),
			HandlerFunc:  labelListHandler},
		route.Route{
			Name:        "LabelDelete",
			Method:      "DELETE",
			Pattern:     "/snapshots/labels/{labelname}",
			Version:     1,
			HandlerFunc: labelDeleteHandler},
		route.Route{
			Name:         "LabelConfigSet",
			Method:       "POST",
			Pattern:      "/snapshots/labels/{labelname}/config",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.LabelSetReq)(nil)),
			ResponseType: utils.GetTypeString((*api.LabelConfigResp)(nil)),
			HandlerFunc:  labelConfigSetHandler},
		route.Route{
			Name:         "LabelConfigReset",
			Method:       "DELETE",
			Pattern:      "/snapshots/labels/{labelname}/config",
			Version:      1,
			RequestType:  utils.GetTypeString((*api.LabelResetReq)(nil)),
			ResponseType: utils.GetTypeString((*api.LabelConfigResp)(nil)),
			HandlerFunc:  labelConfigResetHandler},
	}
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
	registerLabelCreateStepFuncs()
	registerLabelDeleteStepFuncs()
	registerLabelConfigSetStepFuncs()
	registerLabelConfigResetStepFuncs()
	return
}
