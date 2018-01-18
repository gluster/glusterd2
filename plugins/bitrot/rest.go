package bitrot

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/pborman/uuid"
)

func bitrotEnableHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	volName := p["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate volume existence
	volinfo, err := volume.GetVolume(volName)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	// Check if volume is started
	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted.Error(), api.ErrCodeDefault)
		return
	}

	// Check if bitrot is already enabled
	val, exists := volinfo.Options[volume.VkeyFeaturesBitrot]
	if exists && val == "on" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrBitrotAlreadyEnabled.Error(), api.ErrCodeDefault)
		return
	}

	// Enable bitrot-stub
	volinfo.Options[volume.VkeyFeaturesBitrot] = "on"

	/* Enable scrubber daemon (bit-rot.so). The same so acts as bitd and scrubber. With "scrubber" on, it behaves as
	   scrubber othewise bitd */
	volinfo.Options[volume.VkeyFeaturesScrub] = "true"

	// Transaction which starts bitd and scrubber on all nodes.
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	//Lock on Volume Name
	lock, unlock, err := transaction.CreateLockSteps(volName)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = volinfo.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			// Required because bitrot-stub should be enabled on brick side
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  txn.Nodes,
		},

		{
			DoFunc: "bitrot-enable.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	err = txn.Do()
	if err != nil {
		/* TODO: Need to handle failure case. Unlike other daemons,
		 * bitrot daemon is one per node and depends on volfile change.
		 * Need to handle scenarios where bitrot enable is succeeded in
		 * few nodes and failed in few others */
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volName,
		}).Error("failed to enable bitrot")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "bitrot enabled")
}

func bitrotDisableHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	volName := p["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate volume existence
	volinfo, err := volume.GetVolume(volName)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	// Check if bitrot is already disabled
	val, exists := volinfo.Options[volume.VkeyFeaturesBitrot]
	if !exists || val == "off" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrBitrotAlreadyDisabled.Error(), api.ErrCodeDefault)
		return
	}

	// Disable bitrot-stub
	volinfo.Options[volume.VkeyFeaturesBitrot] = "off"
	// Disable scrub by updating volinfo Options
	volinfo.Options[volume.VkeyFeaturesScrub] = "false"

	// Transaction which stop bitd and scrubber on all nodes.
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	//Lock on Volume Name
	lock, unlock, err := transaction.CreateLockSteps(volName)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = volinfo.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			// Required because bitrot-stub should be enabled on brick side
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "bitrot-disable.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	err = txn.Do()
	if err != nil {
		// TODO: Handle rollback
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volName,
		}).Error("failed to disable bitrot")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "bitrot Disable")
}
