package bitrot

import (
	"net/http"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	bitrotapi "github.com/gluster/glusterd2/plugins/bitrot/api"
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
	if volume.IsBitrotEnabled(volinfo) {
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
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Bitrot enabled successfully")
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
	if !volume.IsBitrotEnabled(volinfo) {
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

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Bitrot disabled successfully")
}

func bitrotScrubOndemandHandler(w http.ResponseWriter, r *http.Request) {
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

	// Check if bitrot is disabled
	if !volume.IsBitrotEnabled(volinfo) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrBitrotNotEnabled.Error(), api.ErrCodeDefault)
		return
	}

	// Transaction which starts scrubber on demand.
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

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
			DoFunc: "bitrot-scrubondemand.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("volname", volName)

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volName,
		}).Error("failed to start scrubber")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Scrubber started successfully")
}

func bitrotScrubStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	volName := p["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate volume existence
	volinfo, err := volume.GetVolume(volName)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound,
			errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	// Check if volume is started
	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest,
			errors.ErrVolNotStarted.Error(), api.ErrCodeDefault)
		return
	}

	// Check if bitrot is disabled
	if !volume.IsBitrotEnabled(volinfo) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest,
			errors.ErrBitrotNotEnabled.Error(), api.ErrCodeDefault)
		return
	}

	// Transaction which gets scrubber status.
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	//Lock on Volume Name
	lock, unlock, err := transaction.CreateLockSteps(volName)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError,
			err.Error(), api.ErrCodeDefault)
		return
	}

	// Some nodes may not be up, which is okay.
	txn.DontCheckAlive = true
	txn.DisableRollback = true

	txn.Nodes = volinfo.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "bitrot-scrubstatus.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("volname", volName)

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volName,
		}).Error("failed to get scrubber status")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError,
			err.Error(), api.ErrCodeDefault)
		return
	}

	result, err := createScrubStatusResp(txn.Ctx, volinfo)
	if err != nil {
		errMsg := "Failed to aggregate scrub status results from multiple nodes."
		logger.WithField("error", err.Error()).Error("bitrotScrubStatusHandler:" + errMsg)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError,
			errMsg, api.ErrCodeDefault)
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
	resp.Frequency, exists = volinfo.Options[volume.VkeyScrubFrequency]
	if !exists {
		// If not available in Options, it's not set. Use default value
		opt, err := xlator.FindOption(volume.VkeyScrubFrequency)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err.Error(),
				"volname": volinfo.Name,
			}).Error("failed to get scrub-freq option")
			return &resp, err
		}
		resp.Frequency = opt.DefaultValue
	}

	resp.Throttle, exists = volinfo.Options[volume.VkeyScrubThrottle]
	if !exists {
		// If not available in Options, it's not set. Use default value
		opt, err := xlator.FindOption(volume.VkeyScrubThrottle)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err.Error(),
				"volname": volinfo.Name,
			}).Error("failed to get scrub-throttle option")
			return &resp, err
		}
		resp.Throttle = opt.DefaultValue
	}
	//Bitd log file
	bitrotDaemon, err := newBitd()
	if err != nil {
		log.WithError(err).Error("Failed to create Bitd instance")
		return &resp, err
	}
	resp.BitdLogFile = bitrotDaemon.logfile

	//Scrub log file
	scrubDaemon, err := newScrubd()
	if err != nil {
		log.WithError(err).Error("Failed to create Scrubd instance")
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
			log.WithError(err).Error("strconv of ScrubRunning failed")
			return &resp, err
		}
		if scrubRunning == 1 {
			resp.State = "Active (In Progress)"
		}
		resp.Nodes = append(resp.Nodes, tmp)
	}

	return &resp, nil
}
