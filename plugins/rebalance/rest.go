package rebalance

import (
	"fmt"
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
	// GfDefragStatusNotStarted should be set only for a volume in which rebalance process is not started
	GfDefragStatusNotStarted RebalanceStatus = iota
	// GfDefragStatusStarted should be set only for a volume that has been just started rebalance process
	GfDefragStatusStarted
	// GfDefragStatusStopped should be set only for a volume that has been just stopped rebalance process
	GfDefragStatusStopped
	// GfDefragStatusComplete should be set only for a volume that the rebalance process is completed
	GfDefragStatusComplete
	// GfDefragStatusFailed should be set only for a volume that are failed to run rebalance process
	GfDefragStatusFailed
	// GfDefragStatusLayoutFixStarted should be set only for a volume that has been just started rebalance fix-layout
	GfDefragStatusLayoutFixStarted
	// GfDefragStatusLayoutFixStopped should be set only for a volume that has been just stopped rebalance fix-layout
	GfDefragStatusLayoutFixStopped
	// GfDefragStatusLayoutFixComplete should be set only for a volume that the rebalance fix-layout is completed
	GfDefragStatusLayoutFixComplete
	// GfDefragStatusLayoutFixFailed should be set only for a volume that are failed to run rebalance fix-layout
	GfDefragStatusLayoutFixFailed
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
	ElapsedTime       uint64
	SkippedFiles      uint64
	TimeLeft          uint64
	CommitHash        uint64
}

func createRebalanceInfo(rebalanceinfo *RebalanceInfo) (*RebalanceInfo, error) {
	v := new(RebalanceInfo)
	v.Volname = rebalanceinfo.Volname
	v.RebalanceID = uuid.NewRandom()
	v.Status = rebalanceinfo.Status
	v.CommitHash = rebalanceinfo.CommitHash
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
	transaction.RegisterStepFunc(storeRebalanceDetails, "rebal-start.StoreVolume")
}

func registerRebalanceStopStepFuncs() {
	transaction.RegisterStepFunc(StopRebalance, "rebal-stop.Commit")
}

func registerRebalanceStatusStepFuncs() {
	transaction.RegisterStepFunc(StatusRebalance, "rebal-status.Commit")
}

func rebalanceStart(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	//collect inputs from url
	volname := mux.Vars(r)["volname"]

	rebalanceinfo := new(RebalanceInfo)

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

	if vol.DistCount == 1 {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Volume is not distributed volume or contain only 1 brick",
			api.ErrCodeDefault)
		return
	}

	/*Check volume is a tier volume or not
	if vol.Type == volume.Tier {
	        restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Rebalance operations are not supported on a tier volume",
	                                api.ErrCodeDefault)
	        return
	}*/

	// Check rebalance is already in process
	if rebalanceinfo.Status == GfDefragStatusStarted {
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
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "rebal-start.StoreVolume",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	rebalanceinfo.Volname = volname
	rebalanceinfo.Status = GfDefragStatusStarted
	setcommithash(rebalanceinfo)
	rebal, err := createRebalanceInfo(rebalanceinfo)
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
	rebalanceinfo := new(RebalanceInfo)

	//collect inputs from url
	volname := mux.Vars(r)["volname"]

	//Validate rebalance command
	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusNotFound, "Invalid volume", api.ErrCodeDefault)
		return
	}

	if vol.DistCount == 1 {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Volume is not distributed volume or contain only 1 brick",
			api.ErrCodeDefault)
		return
	}

	// Check remove brick operation is running
	//TODO

	//A simple transaction to stop rebalance
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
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	rebalanceinfo.Volname = volname
	rebalanceinfo.Status = GfDefragStatusStopped
	rebal, err := createRebalanceInfo(rebalanceinfo)
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

	e := AddOrUpdateFunc(rebal)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}
	txn.Ctx.Logger().WithField("volname", rebal.Volname).Info("rebalance stopped")
	restutils.SendHTTPResponse(r.Context(), w, http.StatusOK, "Rebalance Stop")
}
func rebalanceStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	rebalanceinfo := new(RebalanceInfo)

	//collect inputs from url
	volname := mux.Vars(r)["volname"]

	//Validate rebalance command
	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusNotFound, "Invalid volume", api.ErrCodeDefault)
		return
	}

	if vol.DistCount == 1 {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Volume is not distributed volume or contain only 1 brick",
			api.ErrCodeDefault)
		return
	}

	rebal, err := GetRebalanceInfo(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err.Error(), api.ErrCodeDefault)
		return
	}

	if rebal.Status == GfDefragStatusNotStarted {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, "Rebalance process is not started on particular volume",
			api.ErrCodeDefault)
		return
	}

	// A simple transaction to get rebalance status
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	rebalanceinfo.Volname = volname
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "rebal-status.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	// TODO: list all nodes where the process is running
	// Currently gives only localhost
	NodeID := gdctx.MyUUID
	rebalinfo, err := GetRebalanceDetails(rebal)
	if err != nil {
		logger.WithError(err).Error("failed to get Rebalance info")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	txn.Ctx.Logger().WithField("volname", rebalanceinfo.Volname).Info("rebalance status")
	restutils.SendHTTPResponse(r.Context(), w, http.StatusOK, "Rebalance Status:"+rebalanceinfo.Volname+":success")
	restutils.SendHTTPResponse(r.Context(), w, http.StatusOK, fmt.Sprintf("Running on node:%s", NodeID))
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, rebalinfo)
}
