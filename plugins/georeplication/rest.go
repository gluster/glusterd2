package georeplication

import (
	errs "errors"
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func validateMasterAndSlaveIDFormat(w http.ResponseWriter, masteridRaw string, slaveidRaw string) (uuid.UUID, uuid.UUID, error) {
	// Validate UUID format of Master and Slave Volume ID
	masterid := uuid.Parse(masteridRaw)
	if masterid == nil {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Invalid Master Volume ID")
		return nil, nil, errs.New("Invalid Master Volume ID")
	}

	// Validate UUID format of Slave Volume ID
	slaveid := uuid.Parse(slaveidRaw)
	if slaveid == nil {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Invalid Slave Volume ID")
		return nil, nil, errs.New("Invalid Slave Volume ID")
	}

	return masterid, slaveid, nil
}

func georepStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["masterid"]
	slaveidRaw := p["slaveid"]

	reqID, logger := restutils.GetReqIDandLogger(r)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
			return
		}
		restutils.SendHTTPResponse(w, http.StatusOK, []Session{})
		return
	}

	// Get Volume info, which is required to get the Bricks list
	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}

	// Transaction which updates the Geo-rep session
	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(masterid.String() + slaveid.String())
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO: Transaction step function for setting Volume Options

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "georeplication-status",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}
	txn.Ctx.Set("geosession", geoSession)

	rtxn, e := txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":    e.Error(),
			"masterid": masterid,
			"slaveid":  slaveid,
		}).Error("failed to create geo-replication session")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	// Aggregate the results
	result, err := aggregateGsyncdStatus(rtxn, txn.Nodes)
	if err != nil {
		errMsg := "Failed to aggregate gsyncd status results from multiple nodes."
		logger.WithField("error", err.Error()).Error("gsyncdStatusHandler:" + errMsg)
		restutils.SendHTTPError(w, http.StatusInternalServerError, errMsg)
		return
	}

	for _, b := range vol.Bricks {
		// Set default values to all status fields, If a node or worker is down and
		// status not available these default values will be sent back in response
		geoSession.Workers = append(geoSession.Workers, Worker{
			MasterNode:                 b.Hostname,
			MasterNodeID:               b.NodeID.String(),
			MasterBrickPath:            b.Path,
			MasterBrick:                b.NodeID.String() + ":" + b.Path,
			Status:                     "Unknown",
			LastSyncedTime:             "N/A",
			LastSyncedTimeUTC:          "N/A",
			LastEntrySyncedTime:        "N/A",
			SlaveNode:                  "N/A",
			ChangeDetection:            "N/A",
			CheckpointTime:             "N/A",
			CheckpointTimeUTC:          "N/A",
			CheckpointCompleted:        false,
			CheckpointCompletedTime:    "N/A",
			CheckpointCompletedTimeUTC: "N/A",
			MetaOps:                    "0",
			EntryOps:                   "0",
			DataOps:                    "0",
			FailedOps:                  "0",
			CrawlStatus:                "N/A",
		})
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
		geoSession.Workers[idx].ChangeDetection = statusData.ChangeDetection
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
	restutils.SendHTTPResponse(w, http.StatusOK, geoSession)
}

func georepCreateHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["masterid"]
	slaveidRaw := p["slaveid"]

	reqID, logger := restutils.GetReqIDandLogger(r)

	// Same handler for both Create and Update, if HTTP method is POST then
	// it is update request
	updateRequest := false
	if r.Method == http.MethodPost {
		updateRequest = true
	}

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Parse the JSON body to get additional details of request
	var req GeorepCreateRequest
	if err := utils.GetJSONFromRequest(r, &req); err != nil {
		restutils.SendHTTPError(w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error())
		return
	}

	// TODO: Add validation for POST fields

	// Check if Master volume exists and Matches with passed Volume ID
	vol, e := volume.GetVolume(req.MasterVol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}

	// Check if Master Volume ID from store matches the input Master Volume ID
	if !uuid.Equal(vol.ID, masterid) {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Master volume ID doesn't match")
		return
	}

	// Fetch existing session details from Store, if same
	// session exists then return error
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err == nil {
		if !updateRequest {
			restutils.SendHTTPError(w, http.StatusConflict, "Session already exists")
			return
		}
	} else {
		// Continue only if NotFound error, return if other errors like
		// error while fetching from store or JSON marshal errors
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Fail if update request received for non existing Geo-rep session
		if updateRequest {
			restutils.SendHTTPError(w, http.StatusBadRequest, "geo-replication session not found")
			return
		}

		geoSession = new(Session)
		geoSession.MasterID = masterid
		geoSession.SlaveID = slaveid
	}

	// If Update request then it is not allowed to change master volume,
	// slave volname, Master Volume name and ID is already verified earlier
	// only validate for slave vol name here
	if updateRequest && geoSession.SlaveVol != req.SlaveVol {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Slave Volume name can't be modified")
		return
	}

	// Set/Update the input details
	geoSession.MasterVol = req.MasterVol
	geoSession.SlaveVol = req.SlaveVol
	geoSession.SlaveHosts = req.SlaveHosts
	geoSession.SlaveUser = req.SlaveUser

	// Set Slave user as root if request doesn't contains user name
	if geoSession.SlaveUser == "" {
		geoSession.SlaveUser = "root"
	}

	geoSession.Workers = []Worker{}

	// Transaction which updates the Geo-rep session
	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO: Transaction step function for setting Volume Options

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "georeplication-create.Commit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}
	txn.Ctx.Set("geosession", geoSession)

	_, e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":    e.Error(),
			"masterid": masterid,
			"slaveid":  slaveid,
		}).Error("failed to create geo-replication session")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	geoSession.Status = georepStatusCreated

	e = addOrUpdateSession(geoSession)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	respCode := http.StatusCreated
	if updateRequest {
		respCode = http.StatusOK
	}

	restutils.SendHTTPResponse(w, respCode, geoSession)
}

func georepStartHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["masterid"]
	slaveidRaw := p["slaveid"]

	reqID, logger := restutils.GetReqIDandLogger(r)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		// Continue only if NotFound error, return if other errors like
		// error while fetching from store or JSON marshal errors
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
			return
		}
		restutils.SendHTTPError(w, http.StatusBadRequest, "geo-replication session not found")
		return
	}

	if geoSession.Status == georepStatusStarted {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "session already started")
		return
	}

	// TODO: Check Volume is in Started State
	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}

	if vol.Status != volume.VolStarted {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "master volume not started")
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	// TODO: change the lock key
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "georeplication-start.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("masterid", masterid.String())
	txn.Ctx.Set("slaveid", slaveid.String())

	_, e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":    e.Error(),
			"masterid": masterid,
			"slaveid":  slaveid,
		}).Error("failed to start geo-replication session")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	geoSession.Status = georepStatusStarted

	e = addOrUpdateSession(geoSession)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, geoSession)
}

func georepPauseHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	masteridRaw := p["masterid"]
	slaveidRaw := p["slaveid"]

	reqID, logger := restutils.GetReqIDandLogger(r)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		// Continue only if NotFound error, return if other errors like
		// error while fetching from store or JSON marshal errors
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
			return
		}
		restutils.SendHTTPError(w, http.StatusBadRequest, "geo-replication session not found")
		return
	}

	if geoSession.Status == georepStatusPaused {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "session already in pause state")
		return
	}

	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	// TODO: change the lock key
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc:   "georeplication-pause.Commit",
			UndoFunc: "georeplication-pause.Undo",
			Nodes:    txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("masterid", masterid.String())
	txn.Ctx.Set("slaveid", slaveid.String())

	_, e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":    e.Error(),
			"masterid": masterid,
			"slaveid":  slaveid,
		}).Error("failed to start geo-replication session")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	geoSession.Status = georepStatusPaused

	e = addOrUpdateSession(geoSession)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, geoSession)
}

func georepResumeHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	masteridRaw := p["masterid"]
	slaveidRaw := p["slaveid"]

	reqID, logger := restutils.GetReqIDandLogger(r)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		// Continue only if NotFound error, return if other errors like
		// error while fetching from store or JSON marshal errors
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
			return
		}
		restutils.SendHTTPError(w, http.StatusBadRequest, "geo-replication session not found")
		return
	}

	if geoSession.Status != georepStatusPaused {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "session not in paused state")
		return
	}

	// TODO: Check Volume is in Started State
	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}

	if vol.Status != volume.VolStarted {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "master volume not started")
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	// TODO: change the lock key
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc:   "georeplication-resume.Commit",
			UndoFunc: "georeplication-resume.Undo",
			Nodes:    txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("masterid", masterid.String())
	txn.Ctx.Set("slaveid", slaveid.String())

	_, e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":    e.Error(),
			"masterid": masterid,
			"slaveid":  slaveid,
		}).Error("failed to resume geo-replication session")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	geoSession.Status = georepStatusStarted

	e = addOrUpdateSession(geoSession)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, geoSession)
}

func georepStopHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	masteridRaw := p["masterid"]
	slaveidRaw := p["slaveid"]

	reqID, logger := restutils.GetReqIDandLogger(r)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		// Continue only if NotFound error, return if other errors like
		// error while fetching from store or JSON marshal errors
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
			return
		}
		restutils.SendHTTPError(w, http.StatusBadRequest, "geo-replication session not found")
		return
	}

	if geoSession.Status == georepStatusStopped {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "session already stopped")
		return
	}

	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	// TODO: change the lock key
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc:   "georeplication-stop.Commit",
			UndoFunc: "georeplication-stop.Undo",
			Nodes:    txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("masterid", masterid.String())
	txn.Ctx.Set("slaveid", slaveid.String())

	_, e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":    e.Error(),
			"masterid": masterid,
			"slaveid":  slaveid,
		}).Error("failed to stop geo-replication session")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	geoSession.Status = georepStatusStopped

	e = addOrUpdateSession(geoSession)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, geoSession)
}

func georepDeleteHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	masteridRaw := p["masterid"]
	slaveidRaw := p["slaveid"]

	reqID, logger := restutils.GetReqIDandLogger(r)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		// Continue only if NotFound error, return if other errors like
		// error while fetching from store or JSON marshal errors
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
			return
		}
		restutils.SendHTTPError(w, http.StatusBadRequest, "geo-replication session not found")
		return
	}

	if geoSession.Status == georepStatusStarted {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "session already started")
		return
	}

	// TODO: Check Volume is in Started State
	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	// TODO: change the lock key
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc:   "georeplication-delete.Commit",
			UndoFunc: "georeplication-delete.Undo",
			Nodes:    txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("masterid", masterid.String())
	txn.Ctx.Set("slaveid", slaveid.String())

	_, e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":    e.Error(),
			"masterid": masterid,
			"slaveid":  slaveid,
		}).Error("failed to delete geo-replication session")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	e = deleteSession(geoSession)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, geoSession)
}

func georepConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Geo-replication Config Get")
}

func georepConfigSetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Geo-replication Config Set")
}

func georepConfigResetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Geo-replication Config Reset")
}

func georepCheckpointSetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Geo-replication Checkpoint Set")
}

func georepCheckpointResetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Geo-replication Checkpoint Reset")
}

func georepCheckpointGetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Geo-replication Checkpoint Status")
}
