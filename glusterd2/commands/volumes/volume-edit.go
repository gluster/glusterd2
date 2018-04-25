package volumecommands

import (
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func volumeEditHandler(w http.ResponseWriter, r *http.Request) {

	p := mux.Vars(r)
	volname := p["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	var req api.VolEditReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
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

	for key := range req.Metadata {
		if strings.HasPrefix(key, "_") {
			logger.WithField("key", key).Error("Key names starting with '_' are restricted in metadata field")
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Key names starting with '_' are restricted in metadata field")
			return
		}
	}

	v.Metadata = req.Metadata

	// Transaction which starts self heal daemon on all nodes with atleast one brick.
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	txn.Nodes = v.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}

	if err := txn.Ctx.Set("volinfo", v); err != nil {
		logger.WithError(err).WithField("key", "volinfo").Error("failed to set key in transaction context")
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
