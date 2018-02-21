package georeplication

import (
	"context"
	"encoding/json"
	errs "errors"
	"fmt"
	"net/http"
	"os/exec"
	"path"

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
	config "github.com/spf13/viper"
)

// newGeorepSession creates new instance of GeorepSession
func newGeorepSession(mastervolid uuid.UUID, remotevolid uuid.UUID, req georepapi.GeorepCreateReq) *georepapi.GeorepSession {
	remoteUser := req.RemoteUser
	if req.RemoteUser == "" {
		remoteUser = "root"
	}
	remotehosts := make([]georepapi.GeorepRemoteHost, len(req.RemoteHosts))
	for idx, s := range req.RemoteHosts {
		remotehosts[idx].NodeID = uuid.Parse(s.NodeID)
		remotehosts[idx].Hostname = s.Hostname
	}

	return &georepapi.GeorepSession{
		MasterID:    mastervolid,
		RemoteID:    remotevolid,
		MasterVol:   req.MasterVol,
		RemoteVol:   req.RemoteVol,
		RemoteHosts: remotehosts,
		RemoteUser:  remoteUser,
		Status:      georepapi.GeorepStatusCreated,
		Workers:     []georepapi.GeorepWorker{},
		Options:     make(map[string]string),
	}
}

func validateMasterAndRemoteIDFormat(ctx context.Context, w http.ResponseWriter, masteridRaw string, remoteidRaw string) (uuid.UUID, uuid.UUID, error) {
	// Validate UUID format of Master and Remote Volume ID
	masterid := uuid.Parse(masteridRaw)
	if masterid == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid Master Volume ID", api.ErrCodeDefault)
		return nil, nil, errs.New("Invalid Master Volume ID")
	}

	// Validate UUID format of Remote Volume ID
	remoteid := uuid.Parse(remoteidRaw)
	if remoteid == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid Remote Volume ID", api.ErrCodeDefault)
		return nil, nil, errs.New("Invalid Remote Volume ID")
	}

	return masterid, remoteid, nil
}

func georepCreateHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	remoteidRaw := p["remotevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Remote Volume ID
	masterid, remoteid, err := validateMasterAndRemoteIDFormat(ctx, w, masteridRaw, remoteidRaw)
	if err != nil {
		return
	}

	if uuid.Equal(masterid, remoteid) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Master and Remote Volume can't be same", api.ErrCodeDefault)
		return
	}

	// Parse the JSON body to get additional details of request
	var req georepapi.GeorepCreateReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	// Required fields are MasterVol, RemoteHosts and RemoteVol
	if req.MasterVol == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Master volume name is required field", api.ErrCodeDefault)
		return
	}

	if len(req.RemoteHosts) == 0 {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Atleast one Remote host is required", api.ErrCodeDefault)
		return
	}

	if req.RemoteVol == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Remote volume name is required field", api.ErrCodeDefault)
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
	sessionExists := false
	geoSession, err := getSession(masterid.String(), remoteid.String())
	if err == nil {
		sessionExists = true
		if !req.Force {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, "Session already exists", api.ErrCodeDefault)
			return
		}
	}

	// Continue only if NotFound error, return if other errors like
	// error while fetching from store or JSON marshal errors
	if err != nil {
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
			return
		}
	}

	// Initialize only if New Session
	if !sessionExists {
		geoSession = newGeorepSession(masterid, remoteid, req)

		// Set Required Geo-rep Configurations
		geoSession.Options["gluster-rundir"] = path.Join(config.GetString("rundir"), "gluster")
		geoSession.Options["glusterd-workdir"] = config.GetString("localstatedir")
		geoSession.Options["gluster-logdir"] = path.Join(config.GetString("logdir"), "glusterfs")
	}

	// Transaction which updates the Geo-rep session
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	// Lock on Master Volume name
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	// Required Volume Options
	vol.Options["marker.xtime"] = "on"
	vol.Options["marker.gsync-force-xtime"] = "on"
	vol.Options["changelog.changelog"] = "on"

	// Workaround till {{ volume.id }} added to the marker options table
	vol.Options["marker.volume-uuid"] = vol.ID.String()

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "georeplication-create.Commit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}

	if err = txn.Ctx.Set("geosession", geoSession); err != nil {
		logger.WithError(err).Error("failed to set geosession in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	if err = txn.Ctx.Set("volinfo", vol); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"remotevolid": remoteid,
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
	remoteidRaw := p["remotevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Remote Volume ID
	masterid, remoteid, err := validateMasterAndRemoteIDFormat(ctx, w, masteridRaw, remoteidRaw)
	if err != nil {
		return
	}

	// Parse the JSON body to get additional details of request
	var req georepapi.GeorepCommandsReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), remoteid.String())
	if err != nil {
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
			return
		}
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "geo-replication session not found", api.ErrCodeDefault)
		return
	}

	if action == actionStart && geoSession.Status == georepapi.GeorepStatusStarted && !req.Force {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "session already started", api.ErrCodeDefault)
		return
	}

	if action == actionStop && geoSession.Status == georepapi.GeorepStatusStopped && !req.Force {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "session already stopped", api.ErrCodeDefault)
		return
	}

	if action == actionPause && geoSession.Status != georepapi.GeorepStatusStarted && !req.Force {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "session is not in started state", api.ErrCodeDefault)
		return
	}

	if action == actionResume && geoSession.Status != georepapi.GeorepStatusPaused && !req.Force {
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

	if err = txn.Ctx.Set("mastervolid", masterid.String()); err != nil {
		logger.WithError(err).Error("failed to set mastervolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("remotevolid", remoteid.String()); err != nil {
		logger.WithError(err).Error("failed to set remotevolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":       err.Error(),
			"mastervolid": masterid,
			"remotevolid": remoteid,
		}).Error("failed to " + action.String() + " geo-replication session")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
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
	remoteidRaw := p["remotevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Remote Volume ID
	masterid, remoteid, err := validateMasterAndRemoteIDFormat(ctx, w, masteridRaw, remoteidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), remoteid.String())
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

	if err = txn.Ctx.Set("mastervolid", masterid.String()); err != nil {
		logger.WithError(err).Error("failed to set mastervolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("remotevolid", remoteid.String()); err != nil {
		logger.WithError(err).Error("failed to set remotevolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"remotevolid": remoteid,
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
	remoteidRaw := p["remotevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Remote Volume ID
	masterid, remoteid, err := validateMasterAndRemoteIDFormat(ctx, w, masteridRaw, remoteidRaw)
	if err != nil {
		return
	}

	geoSession, err := getSession(masterid.String(), remoteid.String())
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

	if err = txn.Ctx.Set("mastervolid", masterid.String()); err != nil {
		logger.WithError(err).Error("failed to set mastervolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("remotevolid", remoteid.String()); err != nil {
		logger.WithError(err).Error("failed to set remotevolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	e = txn.Do()
	if e != nil {
		// TODO: Handle partial failure if a few glusterd's down

		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"remotevolid": remoteid,
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

	for _, b := range vol.GetBricks() {
		// Set default values to all status fields, If a node or worker is down and
		// status not available these default values will be sent back in response
		geoSession.Workers = append(geoSession.Workers, georepapi.GeorepWorker{
			MasterNode:                 b.Hostname,
			MasterNodeID:               b.NodeID.String(),
			MasterBrickPath:            b.Path,
			MasterBrick:                b.NodeID.String() + ":" + b.Path,
			Status:                     "Unknown",
			LastSyncedTime:             "N/A",
			LastSyncedTimeUTC:          "N/A",
			LastEntrySyncedTime:        "N/A",
			RemoteNode:                 "N/A",
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

	// Iterating and assigning status of each brick and not doing direct
	// assignment. So that order of the workers will be maintained similar
	// to order of bricks in Master Volume
	for idx, w := range geoSession.Workers {
		statusData := (*result)[w.MasterNodeID+":"+w.MasterBrickPath]
		geoSession.Workers[idx].Status = statusData.Status
		geoSession.Workers[idx].LastSyncedTime = statusData.LastSyncedTime
		geoSession.Workers[idx].LastSyncedTimeUTC = statusData.LastSyncedTimeUTC
		geoSession.Workers[idx].LastEntrySyncedTime = statusData.LastEntrySyncedTime
		geoSession.Workers[idx].RemoteNode = statusData.RemoteNode
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

func restartRequiredOnConfigChange(name string) bool {
	// TODO: Check with Gsyncd about restart required or not
	// for now restart gsyncd for all config changes
	return true
}

func checkConfig(name string, value string) error {
	args := []string{
		"config-check",
		name,
	}
	if value != "" {
		args = append(args, "--value", value)
	}
	_, err := exec.Command(gsyncdCommand, args...).Output()
	return err
}

func georepConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	remoteidRaw := p["remotevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Remote Volume ID
	masterid, remoteid, err := validateMasterAndRemoteIDFormat(ctx, w, masteridRaw, remoteidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), remoteid.String())
	if err != nil {
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
			return
		}
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "geo-replication session not found", api.ErrCodeDefault)
		return
	}

	// Run Local gsyncd and get all configs and its default values
	args := []string{
		"config-get",
		geoSession.MasterVol,
		fmt.Sprintf("%s@%s::%s", geoSession.RemoteUser, geoSession.RemoteHosts[0], geoSession.RemoteVol),
		"--show-defaults",
		"--json",
	}
	out, err := exec.Command(gsyncdCommand, args...).Output()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":       err.Error(),
			"mastervolid": masterid,
			"remotevolid": remoteid,
		}).Error("failed to get session configurations")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Failed to get session configurations", api.ErrCodeDefault)
		return
	}

	var opts []georepapi.GeorepOption
	if err = json.Unmarshal(out, &opts); err != nil {
		logger.WithFields(log.Fields{
			"error":       err.Error(),
			"mastervolid": masterid,
			"remotevolid": remoteid,
		}).Error("failed to parse configurations")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Failed to parse configurations", api.ErrCodeDefault)
		return
	}

	// Reset all configurations Value since Gsyncd may return stale data
	// if a old config file exists on disk with stale data(Only happens
	// if Gsyncd is not in Started state)
	for idx, conf := range opts {
		if conf.Modified {
			opts[idx].Modified = false
			opts[idx].Value = opts[idx].DefaultValue
			opts[idx].DefaultValue = ""
		}
	}

	// Now Apply session configurations
	for idx, conf := range opts {
		if val, ok := geoSession.Options[conf.Name]; ok {
			// Gsyncd opt Value is default value
			opts[idx].DefaultValue = opts[idx].Value
			// Add Value from Store
			opts[idx].Value = val
			opts[idx].Modified = true
		}
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, opts)
}

func georepConfigSetHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	remoteidRaw := p["remotevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Remote Volume ID
	masterid, remoteid, err := validateMasterAndRemoteIDFormat(ctx, w, masteridRaw, remoteidRaw)
	if err != nil {
		return
	}

	// Parse the JSON body to get additional details of request
	var req map[string]string
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), remoteid.String())
	if err != nil {
		// Continue only if NotFound error, return if other errors like
		// error while fetching from store or JSON marshal errors
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
			return
		}
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "geo-replication session not found", api.ErrCodeDefault)
		return
	}

	configWillChange := false
	restartRequired := false
	// Validate all config names and values
	for k, v := range req {
		val, ok := geoSession.Options[k]
		if (ok && v != val) || !ok {
			configWillChange = true
			err = checkConfig(k, v)
			if err != nil {
				restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid Config Name/Value", api.ErrCodeDefault)
				return
			}

			restartRequired = restartRequiredOnConfigChange(k)
		}
	}

	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	// If No configurations changed
	if !configWillChange {
		restutils.SendHTTPResponse(ctx, w, http.StatusOK, geoSession)
		return
	}

	// No Restart required if Georep session not running
	if geoSession.Status != georepapi.GeorepStatusStarted {
		restartRequired = false
	}

	for k, v := range req {
		geoSession.Options[k] = v
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	// TODO: change the lock key
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "georeplication-configset.Commit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "georeplication-configfilegen.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	if err = txn.Ctx.Set("mastervolid", masterid.String()); err != nil {
		logger.WithError(err).Error("failed to set mastervolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("remotevolid", remoteid.String()); err != nil {
		logger.WithError(err).Error("failed to set remotevolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("session", geoSession); err != nil {
		logger.WithError(err).Error("failed to set geosession in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("restartRequired", restartRequired); err != nil {
		logger.WithError(err).Error("failed to set restartrequired in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"remotevolid": remoteid,
		}).Error("failed to update geo-replication session config")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, geoSession.Options)
}

func georepConfigResetHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	remoteidRaw := p["remotevolid"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Validate UUID format of Master and Remote Volume ID
	masterid, remoteid, err := validateMasterAndRemoteIDFormat(ctx, w, masteridRaw, remoteidRaw)
	if err != nil {
		return
	}

	// Parse the JSON body to get additional details of request
	var req []string
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), remoteid.String())
	if err != nil {
		// Continue only if NotFound error, return if other errors like
		// error while fetching from store or JSON marshal errors
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
			return
		}
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "geo-replication session not found", api.ErrCodeDefault)
		return
	}

	configWillChange := false
	restartRequired := false
	// Check if config exists, reset can be done only if it is set before
	for _, k := range req {
		_, ok := geoSession.Options[k]
		if ok {
			configWillChange = true

			restartRequired = restartRequiredOnConfigChange(k)
		}
	}

	// If No configurations changed
	if !configWillChange {
		restutils.SendHTTPResponse(ctx, w, http.StatusOK, geoSession.Options)
		return
	}

	// No Restart required if Georep session not running
	if geoSession.Status != georepapi.GeorepStatusStarted {
		restartRequired = false
	}

	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	for _, k := range req {
		delete(geoSession.Options, k)
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	// TODO: change the lock key
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "georeplication-configset.Commit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "georeplication-configfilegen.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	if err = txn.Ctx.Set("mastervolid", masterid.String()); err != nil {
		logger.WithError(err).Error("failed to set mastervolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("remotevolid", remoteid.String()); err != nil {
		logger.WithError(err).Error("failed to set remotevolid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("session", geoSession); err != nil {
		logger.WithError(err).Error("failed to set geosession in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("restartRequired", restartRequired); err != nil {
		logger.WithError(err).Error("failed to set restartrequired in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"remotevolid": remoteid,
		}).Error("failed to update geo-replication session config")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, geoSession.Options)
}

func georepStatusListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessions, err := getSessionList()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, sessions)
}

func georepSSHKeyGenerateHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	volname := p["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Check if Volume exists
	vol, e := volume.GetVolume(volname)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	// Transaction which updates the Geo-rep session
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	// Lock on Master Volume name
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "georeplication-ssh-keygen.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	if err = txn.Ctx.Set("volname", volname); err != nil {
		logger.WithError(err).Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volname,
		}).Error("failed to generate SSH Keys")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	sshkeys, err := getSSHPublicKeys(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, sshkeys)
}

func georepSSHKeyGetHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	volname := p["volname"]

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	sshkeys, err := getSSHPublicKeys(volname)
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volname,
		}).Error("failed to get SSH public Keys")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, sshkeys)
}

func georepSSHKeyPushHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	volname := p["volname"]

	// TODO: Handle non root user
	user := "root"

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Check if Volume exists
	vol, e := volume.GetVolume(volname)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	// Parse the JSON body to get additional details of request
	var sshkeys []georepapi.GeorepSSHPublicKey
	if err := restutils.UnmarshalRequest(r, &sshkeys); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	// Transaction which updates the Geo-rep session
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	// Lock on Master Volume name
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "georeplication-ssh-keypush.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	if err = txn.Ctx.Set("sshkeys", sshkeys); err != nil {
		logger.WithError(err).Error("failed to set sshkeys in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err = txn.Ctx.Set("user", user); err != nil {
		logger.WithError(err).Error("failed to set user in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   e.Error(),
			"volname": volname,
		}).Error("failed to push SSH Keys")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, "SSH Keys added successfully")
}
