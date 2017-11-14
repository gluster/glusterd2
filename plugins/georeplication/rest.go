package georeplication

import (
	errs "errors"
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/api"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

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

func validateMasterAndSlaveIDFormat(w http.ResponseWriter, masteridRaw string, slaveidRaw string) (uuid.UUID, uuid.UUID, error) {
	// Validate UUID format of Master and Slave Volume ID
	masterid := uuid.Parse(masteridRaw)
	if masterid == nil {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Invalid Master Volume ID", api.ErrCodeDefault)
		return nil, nil, errs.New("Invalid Master Volume ID")
	}

	// Validate UUID format of Slave Volume ID
	slaveid := uuid.Parse(slaveidRaw)
	if slaveid == nil {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Invalid Slave Volume ID", api.ErrCodeDefault)
		return nil, nil, errs.New("Invalid Slave Volume ID")
	}

	return masterid, slaveid, nil
}

func georepCreateHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	slaveidRaw := p["slavevolid"]

	reqID, logger := restutils.GetReqIDandLogger(r)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Parse the JSON body to get additional details of request
	var req georepapi.GeorepCreateReq
	if err := utils.GetJSONFromRequest(r, &req); err != nil {
		restutils.SendHTTPError(w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	// Required fields are MasterVol, SlaveHosts and SlaveVol
	if req.MasterVol == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Master volume name is required field", api.ErrCodeDefault)
		return
	}

	if len(req.SlaveHosts) == 0 {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Atleast one Slave host is required", api.ErrCodeDefault)
		return
	}

	if req.SlaveVol == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Slave volume name is required field", api.ErrCodeDefault)
		return
	}

	// Check if Master volume exists and Matches with passed Volume ID
	vol, e := volume.GetVolume(req.MasterVol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	// Check if Master Volume ID from store matches the input Master Volume ID
	if !uuid.Equal(vol.ID, masterid) {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Master volume ID doesn't match", api.ErrCodeDefault)
		return
	}

	// Fetch existing session details from Store, if same
	// session exists then return error
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err == nil {
		restutils.SendHTTPError(w, http.StatusConflict, "Session already exists", api.ErrCodeDefault)
		return
	}

	// Continue only if NotFound error, return if other errors like
	// error while fetching from store or JSON marshal errors
	if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	geoSession = newGeorepSession(masterid, slaveid, req)

	// Transaction which updates the Geo-rep session
	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()

	// Lock on Master Volume name
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	// TODO: Transaction step function for setting Volume Options
	// As a workaround, Set volume options before enabling Geo-rep session

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
			"error":       e.Error(),
			"mastervolid": masterid,
			"slavevolid":  slaveid,
		}).Error("failed to create geo-replication session")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(w, http.StatusCreated, geoSession)
}

func georepStartHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	slaveidRaw := p["slavevolid"]

	reqID, logger := restutils.GetReqIDandLogger(r)

	// Validate UUID format of Master and Slave Volume ID
	masterid, slaveid, err := validateMasterAndSlaveIDFormat(w, masteridRaw, slaveidRaw)
	if err != nil {
		return
	}

	// Fetch existing session details from Store, error if not exists
	geoSession, err := getSession(masterid.String(), slaveid.String())
	if err != nil {
		if _, ok := err.(*ErrGeorepSessionNotFound); !ok {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
			return
		}
		restutils.SendHTTPError(w, http.StatusBadRequest, "geo-replication session not found", api.ErrCodeDefault)
		return
	}

	if geoSession.Status == georepapi.GeorepStatusStarted {
		restutils.SendHTTPError(w, http.StatusConflict, "session already started", api.ErrCodeDefault)
		return
	}

	// Fetch Volume details and check if Volume is in started state
	vol, e := volume.GetVolume(geoSession.MasterVol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	if vol.State != volume.VolStarted {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "master volume not started", api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()

	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
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
	txn.Ctx.Set("mastervolid", masterid.String())
	txn.Ctx.Set("slavevolid", slaveid.String())

	_, e = txn.Do()
	if e != nil {
		logger.WithFields(log.Fields{
			"error":       e.Error(),
			"mastervolid": masterid,
			"slavevolid":  slaveid,
		}).Error("failed to start geo-replication session")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	geoSession.Status = georepapi.GeorepStatusStarted

	e = addOrUpdateSession(geoSession)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, geoSession)
}
