package georeplication

import (
	"context"
	errs "errors"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// newGeorepSession creates new instance of GeorepSession
func newGeorepSession(mastervolid uuid.UUID, slavevolid uuid.UUID, req georepapi.GeorepCreateReq) *georepapi.GeorepSession {
	slaveUser := req.SlaveUser
	if req.SlaveUser == "" {
		slaveUser = "root"
	}
	return &georepapi.GeorepSession{
		MasterID:   mastervolid,
		SlaveID:    slavevolid,
		MasterVol:  req.MasterVol,
		SlaveVol:   req.SlaveVol,
		SlaveHosts: req.SlaveHosts,
		SlaveUser:  slaveUser,
		Status:     georepapi.GeorepStatusCreated,
		Workers:    []georepapi.GeorepWorker{},
		Options:    make(map[string]string),
	}
}

func validateMasterAndSlaveIDFormat(ctx context.Context, w http.ResponseWriter, masteridRaw string, slaveidRaw string) (uuid.UUID, uuid.UUID, error) {
	// Validate UUID format of Master and Slave Volume ID
	masterid := uuid.Parse(masteridRaw)
	if masterid == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid Master Volume ID", api.ErrCodeDefault)
		return nil, nil, errs.New("Invalid Master Volume ID")
	}

	// Validate UUID format of Slave Volume ID
	slaveid := uuid.Parse(slaveidRaw)
	if slaveid == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid Slave Volume ID", api.ErrCodeDefault)
		return nil, nil, errs.New("Invalid Slave Volume ID")
	}

	return masterid, slaveid, nil
}

func georepCreateHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	slaveidRaw := p["slavevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(ctx, w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Parse the JSON body to get additional details of request
	var req georepapi.GeorepCreateReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	// Required fields are MasterVol, SlaveHosts and SlaveVol
	if req.MasterVol == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Master volume name is required field", api.ErrCodeDefault)
		return
	}

	if len(req.SlaveHosts) == 0 {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Atleast one Slave host is required", api.ErrCodeDefault)
		return
	}

	if req.SlaveVol == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Slave volume name is required field", api.ErrCodeDefault)
		return
	}

	// Check if Master volume exists and Matches with passed Volume ID
	vol, e := volume.GetVolume(req.MasterVol)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	// Check if Master Volume ID from store matches the input Master Volume ID
	if !uuid.Equal(vol.ID, masterid) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Master volume ID doesn't match", api.ErrCodeDefault)
		return
	}

	// Fetch existing session details from Store, if same
	// session exists then return error
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err == nil {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "Session already exists", api.ErrCodeDefault)
		return
	}

	// Continue only if NotFound error, return if other errors like
	// error while fetching from store or JSON marshal errors
	if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	geoSession = newGeorepSession(masterid, slaveid, req)

	// Transaction which updates the Geo-rep session
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	// Lock on Master Volume name
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	// TODO: Transaction step function for setting Volume Options
	// As a workaround, Set volume options before enabling Geo-rep session

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "georeplication-create.Commit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}
	txn.Ctx.Set("geosession", geoSession)

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"slavevolid":  slaveid,
		}).Error("failed to create geo-replication session")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, geoSession)
}

func georepActionHandler(w http.ResponseWriter, r *http.Request, action actionType) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	slaveidRaw := p["slavevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(ctx, w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
			return
		}
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "geo-replication session not found", api.ErrCodeDefault)
		return
	}

	if action == actionStart && geoSession.Status == georepapi.GeorepStatusStarted {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "session already started", api.ErrCodeDefault)
		return
	}

	if action == actionStop && geoSession.Status == georepapi.GeorepStatusStopped {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "session already stopped", api.ErrCodeDefault)
		return
	}

	if action == actionPause && geoSession.Status != georepapi.GeorepStatusStarted {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "session is not in started state", api.ErrCodeDefault)
		return
	}

	if action == actionResume && geoSession.Status != georepapi.GeorepStatusPaused {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "session not in paused state", api.ErrCodeDefault)
		return
	}

	// Fetch Volume details and check if Volume is in started state
	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	if action == actionStart && vol.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "master volume not started", api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	doFunc := ""
	stateToSet := ""
	switch action {
	case actionStart:
		doFunc = "georeplication-start.Commit"
		stateToSet = georepapi.GeorepStatusStarted
	case actionPause:
		doFunc = "georeplication-pause.Commit"
		stateToSet = georepapi.GeorepStatusPaused
	case actionResume:
		doFunc = "georeplication-resume.Commit"
		stateToSet = georepapi.GeorepStatusStarted
	case actionStop:
		doFunc = "georeplication-stop.Commit"
		stateToSet = georepapi.GeorepStatusStopped
	default:
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Unknown action", api.ErrCodeDefault)
		return
	}

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: doFunc,
			Nodes:  vol.Nodes(),
		},
		unlock,
	}
	txn.Ctx.Set("mastervolid", masterid.String())
	txn.Ctx.Set("slavevolid", slaveid.String())

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"slavevolid":  slaveid,
		}).Error("failed to " + action.String() + " geo-replication session")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	geoSession.Status = stateToSet

	e = addOrUpdateSession(geoSession)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, geoSession)
}

func georepStartHandler(w http.ResponseWriter, r *http.Request) {
	georepActionHandler(w, r, actionStart)
}

func georepPauseHandler(w http.ResponseWriter, r *http.Request) {
	georepActionHandler(w, r, actionPause)
}

func georepResumeHandler(w http.ResponseWriter, r *http.Request) {
	georepActionHandler(w, r, actionResume)
}

func georepStopHandler(w http.ResponseWriter, r *http.Request) {
	georepActionHandler(w, r, actionStop)
}

func georepDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	slaveidRaw := p["slavevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(ctx, w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
			return
		}
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "geo-replication session not found", api.ErrCodeDefault)
		return
	}

	// Fetch Volume details and check if Volume exists
	_, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	// TODO: Add transaction step to clean xattrs specific to georep session
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "georeplication-delete.Commit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}
	txn.Ctx.Set("mastervolid", masterid.String())
	txn.Ctx.Set("slavevolid", slaveid.String())

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"slavevolid":  slaveid,
		}).Error("failed to delete geo-replication session")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}

func georepStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	slaveidRaw := p["slavevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(ctx, w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
			return
		}
		restutils.SendHTTPResponse(ctx, w, http.StatusOK, []georepapi.GeorepSession{})
		return
	}

	if geoSession.Status != georepapi.GeorepStatusStarted {
		// Reach brick nodes only if Status is Started,
		// else return just the monitor status
		restutils.SendHTTPResponse(ctx, w, http.StatusOK, geoSession)
		return
	}

	// Get Volume info, which is required to get the Bricks list
	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	// Status Transaction
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "georeplication-status.Commit",
			Nodes:  txn.Nodes,
		},
	}

	txn.Ctx.Set("mastervolid", masterid.String())
	txn.Ctx.Set("slavevolid", slaveid.String())

	e = txn.Do()
	if e != nil {
		// TODO: Handle partial failure if a few glusterd's down

		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"slavevolid":  slaveid,
		}).Error("failed to get status of geo-replication session")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	// Aggregate the results
	result, err := aggregateGsyncdStatus(txn.Ctx, txn.Nodes)
	if err != nil {
		errMsg := "Failed to aggregate gsyncd status results from multiple nodes."
		logger.WithField("error", err.Error()).Error("gsyncdStatusHandler:" + errMsg)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, errMsg, api.ErrCodeDefault)
		return
	}

	for _, subvol := range vol.Subvols {
		for _, b := range subvol.Bricks {
			// Set default values to all status fields, If a node or worker is down and
			// status not available these default values will be sent back in response
			geoSession.Workers = append(geoSession.Workers, georepapi.GeorepWorker{
				MasterNode:                 "",
				MasterNodeID:               b.NodeID.String(),
				MasterBrickPath:            b.Path,
				MasterBrick:                b.NodeID.String() + ":" + b.Path,
				Status:                     "Unknown",
				LastSyncedTime:             "N/A",
				LastSyncedTimeUTC:          "N/A",
				LastEntrySyncedTime:        "N/A",
				SlaveNode:                  "N/A",
				CheckpointTime:             "N/A",
				CheckpointTimeUTC:          "N/A",
				CheckpointCompleted:        "N/A",
				CheckpointCompletedTime:    "N/A",
				CheckpointCompletedTimeUTC: "N/A",
				MetaOps:                    "0",
				EntryOps:                   "0",
				DataOps:                    "0",
				FailedOps:                  "0",
				CrawlStatus:                "N/A",
			})
		}
	}

	// Iterating and assigning status of each brick and not doing direct
	// assignment. So that order of the workers will be maintained similar
	// to order of bricks in Master Volume
	for idx, w := range geoSession.Workers {
		statusData := (*result)[w.MasterNodeID+":"+w.MasterBrickPath]
		geoSession.Workers[idx].Status = statusData.Status
		geoSession.Workers[idx].LastSyncedTime = statusData.LastSyncedTime
		geoSession.Workers[idx].LastSyncedTimeUTC = statusData.LastSyncedTimeUTC
		geoSession.Workers[idx].LastEntrySyncedTime = statusData.LastEntrySyncedTime
		geoSession.Workers[idx].SlaveNode = statusData.SlaveNode
		geoSession.Workers[idx].CheckpointTime = statusData.CheckpointTime
		geoSession.Workers[idx].CheckpointTimeUTC = statusData.CheckpointTimeUTC
		geoSession.Workers[idx].CheckpointCompleted = statusData.CheckpointCompleted
		geoSession.Workers[idx].CheckpointCompletedTime = statusData.CheckpointCompletedTime
		geoSession.Workers[idx].CheckpointCompletedTimeUTC = statusData.CheckpointCompletedTimeUTC
		geoSession.Workers[idx].MetaOps = statusData.MetaOps
		geoSession.Workers[idx].EntryOps = statusData.EntryOps
		geoSession.Workers[idx].DataOps = statusData.DataOps
		geoSession.Workers[idx].FailedOps = statusData.FailedOps
		geoSession.Workers[idx].CrawlStatus = statusData.CrawlStatus
	}

	// Send aggregated result back to the client
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, geoSession)
}
