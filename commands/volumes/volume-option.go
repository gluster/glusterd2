package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"
	"github.com/pborman/uuid"

	"github.com/gorilla/mux"
)

// VolOptionRequest represents an incoming request to set volume options
type VolOptionRequest struct {
	Options map[string]string `json:"options"`
}

func updateVolinfoOnOptionChange(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if err := volume.AddOrUpdateVolumeFunc(&volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("failed to store volume info")
		return err
	}

	return nil
}

func registerVolOptionStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-option.UpdateVolinfo", updateVolinfoOnOptionChange},
		{"vol-option.RegenerateVolfiles", generateVolfiles},
		{"vol-option.NotifyVolfileChange", notifyVolfileChange},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeOptionsHandler(w http.ResponseWriter, r *http.Request) {

	p := mux.Vars(r)
	volname := p["volname"]
	reqID, logger := restutils.GetReqIDandLogger(r)

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolNotFound.Error())
		return
	}

	var req VolOptionRequest
	if err := utils.GetJSONFromRequest(r, &req); err != nil {
		restutils.SendHTTPError(w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error())
		return
	}

	if !areOptionNamesValid(req.Options) {
		logger.Error("invalid volume options provided")
		restutils.SendHTTPError(w, http.StatusBadRequest, "invalid volume options provided")
		return
	}

	lock, unlock, err := transaction.CreateLockSteps(volinfo.Name)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()

	// thes txn framework checks if these nodes are online before txn starts
	txn.Nodes = volinfo.Nodes()

	allNodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-option.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-option.RegenerateVolfiles",
			// BUG: Shouldn't be on all nodes ideally. Currently we
			// can't know if it's a brick option or client option.
			// If it's a brick option, the nodes list here should
			// should be only volinfo.Nodes(). Moving client
			// volfiles from disk to store should also be done.
			Nodes: allNodes,
		},
		{
			DoFunc: "vol-option.NotifyVolfileChange",
			Nodes:  allNodes,
		},
		unlock,
	}

	for k, v := range req.Options {
		volinfo.Options[k] = v
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if _, err := txn.Do(); err != nil {
		logger.WithError(err).Error("volume option transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(w, http.StatusConflict, err.Error())
		} else {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	restutils.SendHTTPResponse(w, http.StatusOK, volinfo.Options)
}
