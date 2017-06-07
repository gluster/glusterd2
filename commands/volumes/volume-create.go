package volumecommands

import (
	"errors"
	"net/http"
	"os"

	gderrors "github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

	"github.com/pborman/uuid"
)

// VolCreateRequest defines the parameters for creating a volume in the volume-create command
type VolCreateRequest struct {
	Name         string   `json:"name"`
	Transport    string   `json:"transport,omitempty"`
	ReplicaCount int      `json:"replica,omitempty"`
	Bricks       []string `json:"bricks"`
	Force        bool     `json:"force,omitempty"`
	// Bricks list is ordered (like in glusterd1) and decides which bricks
	// form replica sets.
}

func unmarshalVolCreateRequest(msg *VolCreateRequest, r *http.Request) (int, error) {
	if err := utils.GetJSONFromRequest(r, msg); err != nil {
		return 422, gderrors.ErrJSONParsingFailed
	}

	if msg.Name == "" {
		return http.StatusBadRequest, gderrors.ErrEmptyVolName
	}
	if len(msg.Bricks) <= 0 {
		return http.StatusBadRequest, gderrors.ErrEmptyBrickList
	}
	return 0, nil

}

func createVolinfo(req *VolCreateRequest) (*volume.Volinfo, error) {

	var err error

	v := new(volume.Volinfo)
	v.Options = make(map[string]string)
	v.ID = uuid.NewRandom()
	v.Name = req.Name

	if len(req.Transport) > 0 {
		v.Transport = req.Transport
	} else {
		v.Transport = "tcp"
	}

	if req.ReplicaCount == 0 {
		v.ReplicaCount = 1
	} else {
		v.ReplicaCount = req.ReplicaCount
	}

	if (len(req.Bricks) % v.ReplicaCount) != 0 {
		return nil, errors.New("Invalid number of bricks")
	}

	v.DistCount = len(req.Bricks) / v.ReplicaCount

	switch len(req.Bricks) {
	case 1:
		fallthrough
	case v.DistCount:
		v.Type = volume.Distribute
	case v.ReplicaCount:
		v.Type = volume.Replicate
	default:
		v.Type = volume.DistReplicate
	}

	v.Bricks, err = volume.NewBrickEntriesFunc(req.Bricks, v.Name, v.ID)
	if err != nil {
		return nil, err
	}

	v.Auth = volume.VolAuth{
		Username: uuid.NewRandom().String(),
		Password: uuid.NewRandom().String(),
	}

	return v, nil
}

func validateVolumeCreate(c transaction.TxnCtx) error {

	var req VolCreateRequest
	err := c.Get("req", &req)
	if err != nil {
		return err
	}

	var volinfo volume.Volinfo
	err = c.Get("volinfo", &volinfo)
	if err != nil {
		return err
	}

	// FIXME: Return values of this function are inconsistent and unused
	if _, err = volume.ValidateBrickEntriesFunc(volinfo.Bricks, volinfo.ID, req.Force); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("validateVolumeCreate: failed to validate bricks")
		return err
	}

	return nil
}

func generateVolfiles(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	// Create 'vols' directory.
	err := os.MkdirAll(utils.GetVolumeDir(volinfo.Name), os.ModeDir|os.ModePerm)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("generateVolfiles: failed to create vol directory")
		return err
	}

	// Generate brick volfiles
	for _, b := range volinfo.Bricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}
		if err := volgen.GenerateBrickVolfile(&volinfo, &b); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Debug("generateVolfiles: failed to create brick volfile")
			return err
		}
	}

	// Generate client volfile
	if err := volgen.GenerateClientVolfile(&volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("generateVolfiles: failed to create client volfile")
		return err
	}

	return nil
}

func storeVolume(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if err := volume.AddOrUpdateVolumeFunc(&volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("storeVolume: failed to store volume info")
		return err
	}

	return nil
}

func rollBackVolumeCreate(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	_ = volume.RemoveBrickPaths(volinfo.Bricks)
	return nil
}

func registerVolCreateStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-create.Stage", validateVolumeCreate},
		{"vol-create.Commit", generateVolfiles},
		{"vol-create.Store", storeVolume},
		{"vol-create.Rollback", rollBackVolumeCreate},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

// nodesForVolCreate returns a list of Nodes which volume create touches
func nodesForVolCreate(req *VolCreateRequest) ([]uuid.UUID, error) {
	var nodes []uuid.UUID

	for _, b := range req.Bricks {

		// Bricks specified can have one of the following formats:
		// <peer-uuid>:<brick-path>
		// <ip>:<port>:<brick-path>
		// <ip>:<brick-path>

		host, _, err := utils.ParseHostAndBrickPath(b)
		if err != nil {
			return nil, err
		}

		id := uuid.Parse(host)
		if id == nil {
			// Host specified is IP or IP:port
			id, err = peer.GetPeerIDByAddrF(host)
			if err != nil {
				return nil, err
			}
		}

		nodes = append(nodes, id)
	}
	return nodes, nil
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {
	req := new(VolCreateRequest)
	reqID, logger := restutils.GetReqIDandLogger(r)

	httpStatus, err := unmarshalVolCreateRequest(req, r)
	if err != nil {
		logger.WithError(err).Error("Failed to unmarshal volume request")
		restutils.SendHTTPError(w, httpStatus, err.Error())
		return
	}

	if volume.ExistsFunc(req.Name) {
		restutils.SendHTTPError(w, http.StatusInternalServerError, gderrors.ErrVolExists.Error())
		return
	}

	nodes, err := nodesForVolCreate(req)
	if err != nil {
		logger.WithError(err).Error("could not prepare node list")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	txn, err := (&transaction.SimpleTxn{
		Nodes:    nodes,
		LockKey:  req.Name,
		Stage:    "vol-create.Stage",
		Commit:   "vol-create.Commit",
		Store:    "vol-create.Store",
		Rollback: "vol-create.Rollback",
	}).NewTxn(reqID)
	if err != nil {
		logger.WithError(err).Error("failed to create transaction")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer txn.Cleanup()

	err = txn.Ctx.Set("req", req)
	if err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vol, err := createVolinfo(req)
	if err != nil {
		logger.WithError(err).Error("failed to create volinfo")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = txn.Ctx.Set("volinfo", vol)
	if err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c, err := txn.Do()
	if err != nil {
		logger.WithError(err).Error("volume create transaction failed")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err = c.Get("volinfo", &vol); err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "failed to get volinfo")
		return
	}

	c.Logger().WithField("volname", vol.Name).Info("new volume created")
	restutils.SendHTTPResponse(w, http.StatusCreated, vol)
}
