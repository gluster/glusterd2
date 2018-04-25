package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func registerVolEditMetadataStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-edit-metadata", editVolMetadata},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeEditMetadataHandler(w http.ResponseWriter, r *http.Request) {

	p := mux.Vars(r)
	volname := p["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	var req api.VolEditMetadataReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, err)
		return
	}

	//Lock on Volume Name
	lock, unlock := transaction.CreateLockFuncs(volname)
	// Taking a lock outside the txn as volinfo.Nodes() must also
	// be populated holding the lock.
	if err := lock(ctx); err != nil {
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	defer unlock(ctx)

	//validate volume name
	v, err := volume.GetVolume(volname)
	if err != nil {
		if err == errors.ErrVolNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}

	// Transaction which starts self heal daemon on all nodes with atleast one brick.
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	txn.Nodes = v.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-edit-metadata",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  txn.Nodes,
		},
	}

	if err := txn.Ctx.Set("volname", volname); err != nil {
		logger.WithError(err).WithField("key", "volname").Error("failed to set key in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("metadata", req); err != nil {
		logger.WithError(err).WithField("key", "metadata").Error("failed to set key in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithField("volname", volname).Error("failed to edit metadata")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	resp := createVolumeGetResp(v)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)

}
