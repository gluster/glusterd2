package rebalance

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// RebalanceStatus represents Rebalance Status
type RebalanceStatus uint64

const (
	// Started should be set only for a volume that has been just started rebalance process
	Started RebalanceStatus = iota
	// Inprogress should be set only for a volume that are running rebalance process
	Inprogress
	// Failed should be set only for a volume that are failed to run rebalance process
	Failed
	// Completed should be set only for a volume that the rebalance process is completed
	Completed
	// Stopped should be set only for a volume that has been just stopped rebalance process
	Stopped
)

// RebalanceInfo represents Rebalance details
type RebalanceInfo struct {
	Volname           string
	Status            RebalanceStatus
	RebalanceID       uuid.UUID
	RebalanceFiles    uint64
	RebalanceData     uint64
	LookedupFiles     uint64
	RebalanceFailures uint64
	RebalanceTime     uint64
	SkippedFiles      uint64
	CommitHash        uint64
}

func createRebalanceInfo(rebalance *RebalanceInfo) (*RebalanceInfo, error) {
	v := new(RebalanceInfo)
	v.Volname = rebalance.Volname
	v.RebalanceID = uuid.NewRandom()
	v.Status = rebalance.Status
	v.CommitHash = rebalance.CommitHash
	return v, nil
}

func rebalanceStartInfoResp(v *RebalanceInfo) *api.RebalanceInfo {

	return &api.RebalanceInfo{
		Volname:     v.Volname,
		Status:      api.RebalanceStatus(v.Status),
		RebalanceID: v.RebalanceID,
	}
}

func registerRebalanceStartStepFuncs() {
	transaction.RegisterStepFunc(StartRebalance, "rebal-start.Commit")
}

func registerRebalanceStopStepFuncs() {
	transaction.RegisterStepFunc(StopRebalance, "rebal-stop.Commit")
}

func rebalanceStart(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	//collect inputs from url
	volname := mux.Vars(r)["volname"]

	rebalance := new(RebalanceInfo)

	//Check volname given
	if volname == "" {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Volume not found", api.ErrCodeDefault)
		return
	}

	//Validate rebalance command
	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusNotFound, "Invalid volume", api.ErrCodeDefault)
		return
	}

	if vol.State != volume.VolStarted {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Volume not started", api.ErrCodeDefault)
		return
	}

	if vol.DistCount <= 1 {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Volume is not distribute type or contain only 1 brick",
			api.ErrCodeDefault)
		return
	}

	// TODO: TIER volume is not defined
	/*Check volume is a tier volume or not
	if vol.Type == volume.Tier {
	        restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Rebalance operations are not supported on a tier volume",
	                                api.ErrCodeDefault)
	        return
	}*/

	// Check rebalance is already in process
	if rebalance.Status == Inprogress {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Rebalance is already in progess",
			api.ErrCodeDefault)
		return
	}

	// Check for remove- brick pending
	//TODO

	//A simple transaction to start rebalance
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "rebal-start.Commit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	rebalance.Volname = volname
	rebalance.Status = Started
	setCommitHash(rebalance)
	rebal, err := createRebalanceInfo(rebalance)
	if err != nil {
		logger.WithError(err).Error("failed to create Rebalance info")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Ctx.Set("rinfo", rebal)
	if err != nil {
		logger.WithError(err).Error("failed to set rebalance info in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volname,
		}).Error("failed to start rebalance on volume")
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Get("rinfo", &rebal); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "failed to get rebalanceinfo", api.ErrCodeDefault)
		return
	}

	txn.Ctx.Logger().WithField("volname", rebal.Volname).Info("rebalance started")
	resp := rebalanceStartResp(rebal)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func rebalanceStartResp(v *RebalanceInfo) *api.RebalanceStartResp {
	return (*api.RebalanceStartResp)(rebalanceStartInfoResp(v))
}

func rebalanceStop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	rebalance := new(RebalanceInfo)

	//collect inputs from url
	volname := mux.Vars(r)["volname"]

	//Validate rebalance command
	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusNotFound, "Invalid volume", api.ErrCodeDefault)
		return
	}

	if vol.DistCount <= 1 {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Volume is not distribute type or contain only 1 brick",
			api.ErrCodeDefault)
		return
	}

	// Check rebalance is already in process
	if rebalance.Status == Inprogress {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Rebalance is already in progess",
			api.ErrCodeDefault)
		return
	}

	// Check remove-brick is in progress
	// TODO

	//A simple transaction to start rebalance
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "rebal-stop.Commit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	rebalance.Volname = volname
	rebalance.Status = Stopped
	rebal, err := createRebalanceInfo(rebalance)
	if err != nil {
		logger.WithError(err).Error("failed to create Rebalance info")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Ctx.Set("rinfo", rebal)
	if err != nil {
		logger.WithError(err).Error("failed to set rebalance info in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volname,
		}).Error("failed to stop rebalance on volume")
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	//txn.Ctx.Logger().WithField("volname", rebal.Volname).Info("rebalance stopped")
	restutils.SendHTTPResponse(r.Context(), w, http.StatusOK, "Rebalance Stop")
}
func rebalanceStatus(w http.ResponseWriter, r *http.Request) {
	// Implement the help logic and send response back as below
	restutils.SendHTTPResponse(r.Context(), w, http.StatusOK, "Rebalance Status")
}
