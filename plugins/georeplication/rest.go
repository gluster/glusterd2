package georeplication

import (
	errs "errors"
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	log "github.com/sirupsen/logrus"
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

func georepCreateUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	masteridRaw := p["mastervolid"]
	slaveidRaw := p["slavevolid"]

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
	var req georepapi.GeorepCreateRequest
	if err := utils.GetJSONFromRequest(r, &req); err != nil {
		restutils.SendHTTPError(w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error())
		return
	}

	// Required fields are MasterVol, SlaveHosts and SlaveVol
	if req.MasterVol == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Master volume name is required field")
		return
	}

	if len(req.SlaveHosts) == 0 {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Atleast one Slave host is required")
		return
	}

	if req.SlaveVol == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Slave volume name is required field")
		return
	}

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

		geoSession = new(georepapi.GeorepSession)
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

	geoSession.Workers = []georepapi.GeorepWorker{}
	geoSession.Options = make(map[string]string)

	// Transaction which updates the Geo-rep session
	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()

	// Lock on Master Volume name
	lock, unlock, err := transaction.CreateLockSteps(geoSession.MasterVol)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
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
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	geoSession.Status = georepapi.GeorepStatusCreated

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
