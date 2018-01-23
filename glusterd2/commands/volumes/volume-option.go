package volumecommands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/glusterd2/xlator/options"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"

	log "github.com/sirupsen/logrus"
)

func registerVolOptionStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-option.UpdateVolinfo", storeVolume},
		{"vol-option.NotifyVolfileChange", notifyVolfileChange},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeOptionsGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	volname := mux.Vars(r)["volname"]
	optname := mux.Vars(r)["optname"]

	v, err := volume.GetVolume(volname)
	if err != nil {
		if volname != "all" {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		} else {
			resp := createVolumeOptionsGetResp(nil, optname)
			restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
		}
		return
	}

	resp := createVolumeOptionsGetResp(v, optname)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeOptionsGetResp(v *volume.Volinfo, optname string) *api.VolumeOptionsGetResp {
	return &api.VolumeOptionsGetResp{
		Key:   optname,
		Value: xlator.GlobalOptMap[optname].Value,
	}
}

func setGlobalOption(req api.VolOptionReq) {
	for k, v := range req.Options {
		// For global options update the store and global optmap

		if o, found := xlator.GlobalOptMap[k]; found {
			globalOpt := options.GlobalOption{k, v, o.DefaultValue, o.ValidateType}
			data, _ := json.Marshal(globalOpt)

			if _, err := store.Store.Put(context.TODO(), xlator.GlobalOptsPrefix+k, string(data)); err != nil {
				restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, fmt.Sprint("Failed to update store for global option %s: %s", k, err.Error()), api.ErrCodeDefault)
				return
			}

			// Load global options
			if err := xlator.LoadGlobalOptions(); err != nil {
				log.WithError(err).Warn("Failed to load global options")
			}

			continue
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, fmt.Sprintf("Invalid global option: %s", k), api.ErrCodeDefault)
			continue
		}
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, xlator.GlobalOptMap)
}

func volumeOptionsSetHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	volname := mux.Vars(r)["volname"]
	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		if volname != "all" {
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
			return
		}
	}

	var req api.VolOptionReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	var options map[string]string
	if options, err = expandOptions(req.Options); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err := validateOptions(options); err != nil {
		logger.WithError(err).Error("failed to set volume option")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, fmt.Sprintf("failed to set volume option: %s", err.Error()), api.ErrCodeDefault)
		return
	}

	if err := validateXlatorOptions(req.Options, volinfo); err != nil {
		logger.WithError(err).Error("validation failed")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, fmt.Sprintf("failed to set volume option: %s", err.Error()), api.ErrCodeDefault)
		return
	}

	lock, unlock, err := transaction.CreateLockSteps(volinfo.Name)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	allNodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  allNodes,
		},
		unlock,
	}

	for k, v := range options {
		// TODO: Normalize <graph>.<xlator>.<option> and just
		// <xlator>.<option> to avoid ambiguity and duplication.
		// For example, currently both the following representations
		// will be stored in volinfo:
		// {"afr.eager-lock":"on","gfproxy.afr.eager-lock":"on"}
		volinfo.Options[k] = v
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).Error("volume option transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err.Error(), api.ErrCodeDefault)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		}
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, volinfo.Options)
}
