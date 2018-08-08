package bitrot

import (
	"net/http"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/errors"
	bitrotapi "github.com/gluster/glusterd2/plugins/bitrot/api"
	"github.com/gorilla/mux"

	"github.com/pborman/uuid"
)

func bitrotEnableHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	volname := mux.Vars(r)["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Validate volume existence
	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	// Check if volume is started
	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted)
		return
	}

	// Check if bitrot is already enabled
	if isBitrotEnabled(volinfo) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrBitrotAlreadyEnabled)
		return
	}
	//save volume information for transaction failure scenario
	if err := txn.Ctx.Set("oldvolinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	// Enable bitrot-stub
	volinfo.Options[keyFeaturesBitrot] = "on"

	/* Enable scrubber daemon (bit-rot.so). The same so acts as bitd and scrubber. With "scrubber" on, it behaves as
	   scrubber othewise bitd */
	volinfo.Options[keyFeaturesScrub] = "true"

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Nodes = volinfo.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc:   "vol-option.UpdateVolinfo",
			UndoFunc: "vol-option.UpdateVolinfo.Undo",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
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
	}

	if err = txn.Do(); err != nil {
		/* TODO: Need to handle failure case. Unlike other daemons,
		 * bitrot daemon is one per node and depends on volfile change.
		 * Need to handle scenarios where bitrot enable is succeeded in
		 * few nodes and failed in few others */
		logger.WithError(err).WithField("volname",
			volinfo.Name).Error("failed to enable bitrot")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}

func bitrotDisableHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	volname := mux.Vars(r)["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Validate volume existence
	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	// Check if bitrot is already disabled
	if !isBitrotEnabled(volinfo) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrBitrotAlreadyDisabled)
		return
	}

	// Disable bitrot-stub
	volinfo.Options[keyFeaturesBitrot] = "off"
	// Disable scrub by updating volinfo Options
	volinfo.Options[keyFeaturesScrub] = "false"

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Nodes = volinfo.Nodes()
	txn.Steps = []*transaction.Step{
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
	}

	if err = txn.Do(); err != nil {
		// TODO: Handle rollback
		logger.WithError(err).WithField("volname",
			volinfo.Name).Error("failed to disable bitrot")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}

func bitrotScrubOndemandHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	volname := mux.Vars(r)["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Validate volume existence
	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	// Check if volume is started
	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted)
		return
	}

	// Check if bitrot is disabled
	if !isBitrotEnabled(volinfo) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrBitrotNotEnabled)
		return
	}

	txn.Nodes = volinfo.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "bitrot-scrubondemand.Commit",
			Nodes:  txn.Nodes,
		},
	}
	if err = txn.Ctx.Set("volname", volname); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).WithField("volname",
			volinfo.Name).Error("failed to start scrubber")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}

func bitrotScrubStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	volname := mux.Vars(r)["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Validate volume existence
	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	// Check if volume is started
	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest,
			errors.ErrVolNotStarted)
		return
	}

	// Check if bitrot is disabled
	if !isBitrotEnabled(volinfo) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest,
			errors.ErrBitrotNotEnabled)
		return
	}

	// Some nodes may not be up, which is okay.
	txn.DontCheckAlive = true
	txn.DisableRollback = true

	txn.Nodes = volinfo.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "bitrot-scrubstatus.Commit",
			Nodes:  txn.Nodes,
		},
	}
	if err = txn.Ctx.Set("volname", volname); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).WithField("volname",
			volinfo.Name).Error("failed to get scrubber status")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	result, err := createScrubStatusResp(txn.Ctx, volinfo)
	if err != nil {
		errMsg := "failed to aggregate scrub status results from multiple nodes"
		logger.WithError(err).WithField("volname",
			volinfo.Name).Error("bitrotScrubStatusHandler:" + errMsg)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, errMsg)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, result)
}

func createScrubStatusResp(ctx transaction.TxnCtx, volinfo *volume.Volinfo) (*bitrotapi.ScrubStatus, error) {

	var resp bitrotapi.ScrubStatus
	var exists bool
	// Fill generic info which are same for each node
	resp.Volume = volinfo.Name
	resp.State = "Active (Idle)"
	resp.Frequency, exists = volinfo.Options[keyScrubFrequency]
	if !exists {
		// If not available in Options, it's not set. Use default value
		opt, err := xlator.FindOption(keyScrubFrequency)
		if err != nil {
			ctx.Logger().WithError(err).WithField("volname",
				volinfo.Name).Error("failed to get scrub-freq option")
			return &resp, err
		}
		resp.Frequency = opt.DefaultValue
	}

	resp.Throttle, exists = volinfo.Options[keyScrubThrottle]
	if !exists {
		// If not available in Options, it's not set. Use default value
		opt, err := xlator.FindOption(keyScrubThrottle)
		if err != nil {
			ctx.Logger().WithError(err).WithField("volname",
				volinfo.Name).Error("failed to get scrub-throttle option")
			return &resp, err
		}
		resp.Throttle = opt.DefaultValue
	}
	//Bitd log file
	bitrotDaemon, err := newBitd()
	if err != nil {
		ctx.Logger().WithError(err).WithField("volname",
			volinfo.Name).Error("Failed to create Bitd instance")
		return &resp, err
	}
	resp.BitdLogFile = bitrotDaemon.logfile

	//Scrub log file
	scrubDaemon, err := newScrubd()
	if err != nil {
		ctx.Logger().WithError(err).WithField("volname",
			volinfo.Name).Error("Failed to create Scrubd instance")
		return &resp, err
	}
	resp.ScrubLogFile = scrubDaemon.logfile

	// Loop over each node that make up the volume and aggregate result
	// of scrub status
	for _, node := range volinfo.Nodes() {
		var tmp bitrotapi.ScrubNodeInfo
		err := ctx.GetNodeResult(node, scrubStatusTxnKey, &tmp)
		if err != nil {
			// skip if we do not have information
			continue
		}

		scrubRunning, err := strconv.Atoi(tmp.ScrubRunning)
		if err != nil {
			ctx.Logger().WithError(err).WithField("volname",
				volinfo.Name).Error("strconv of ScrubRunning failed")
			return &resp, err
		}
		if scrubRunning == 1 {
			resp.State = "Active (In Progress)"
		}
		resp.Nodes = append(resp.Nodes, tmp)
	}

	return &resp, nil
}
