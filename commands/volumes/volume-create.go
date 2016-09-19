package volumecommands

import (
	"errors"
	"net/http"

	gderrors "github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

func unmarshalVolCreateRequest(msg *volume.VolCreateRequest, r *http.Request) (int, error) {
	e := utils.GetJSONFromRequest(r, msg)
	if e != nil {
		log.WithField("error", e).Error("Failed to parse the JSON Request")
		return 422, gderrors.ErrJSONParsingFailed
	}

	if msg.Name == "" {
		log.Error("Volume name is empty")
		return http.StatusBadRequest, gderrors.ErrEmptyVolName
	}
	if len(msg.Bricks) <= 0 {
		log.WithField("volume", msg.Name).Error("Brick list is empty")
		return http.StatusBadRequest, gderrors.ErrEmptyBrickList
	}
	return 0, nil

}

func createVolinfo(msg *volume.VolCreateRequest) (*volume.Volinfo, error) {
	vol, err := volume.NewVolumeEntry(msg)
	if err != nil {
		return nil, err
	}
	vol.Bricks, err = volume.NewBrickEntriesFunc(msg.Bricks)
	if err != nil {
		return nil, err
	}
	return vol, nil
}

func validateVolumeCreate(c transaction.TxnCtx) error {
	var req volume.VolCreateRequest
	e := c.Get("req", &req)
	if e != nil {
		return errors.New("failed to get request from context")
	}

	if volume.ExistsFunc(req.Name) {
		c.Logger().WithField("volume", req.Name).Error("volume already exists")
		return gderrors.ErrVolExists
	}

	vol, err := createVolinfo(&req)
	if err != nil {
		return err
	}

	_, err = volume.ValidateBrickEntriesFunc(vol.Bricks, vol.ID, req.Force)
	if err != nil {
		return err
	}

	// Store volinfo for later usage
	e = c.Set("volinfo", vol)

	return e
}

func generateVolfiles(c transaction.TxnCtx) error {
	var vol volume.Volinfo
	e := c.Get("volinfo", &vol)
	if e != nil {
		return errors.New("failed to get volinfo from context")
	}

	// Creating client and server volfile
	e = volgen.GenerateVolfileFunc(&vol)
	if e != nil {
		c.Logger().WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("failed to generate volfile")
		return e
	}
	return nil
}

func storeVolume(c transaction.TxnCtx) error {
	var vol volume.Volinfo
	e := c.Get("volinfo", &vol)
	if e != nil {
		return errors.New("failed to get volinfo from context")
	}

	e = volume.AddOrUpdateVolumeFunc(&vol)
	if e != nil {
		c.Logger().WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("Failed to create volume")
		return e
	}

	log.WithField("volume", vol.Name).Debug("new volume added")
	return nil
}

func rollBackVolumeCreate(c transaction.TxnCtx) error {
	var vol volume.Volinfo
	e := c.Get("volinfo", &vol)
	if e != nil {
		return errors.New("failed to get volinfo from context")
	}

	_ = volume.RemoveBrickPaths(vol.Bricks)

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
func nodesForVolCreate(req *volume.VolCreateRequest) ([]uuid.UUID, error) {
	var nodes []uuid.UUID

	for _, b := range req.Bricks {
		addr, _, err := utils.ParseHostAndBrickPath(b)
		if err != nil {
			return nil, err
		}
		id, err := peer.GetPeerIDByAddrF(addr)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, id)
	}
	return nodes, nil
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {
	req := new(volume.VolCreateRequest)

	httpStatus, e := unmarshalVolCreateRequest(req, r)
	if e != nil {
		rest.SendHTTPError(w, httpStatus, e.Error())
		return
	}

	nodes, e := nodesForVolCreate(req)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	txn, e := (&transaction.SimpleTxn{
		Nodes:    nodes,
		LockKey:  req.Name,
		Stage:    "vol-create.Stage",
		Commit:   "vol-create.Commit",
		Store:    "vol-create.Store",
		Rollback: "vol-create.Rollback",
		LogFields: &log.Fields{
			"reqid": uuid.NewRandom().String(),
		},
	}).NewTxn()
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	defer txn.Cleanup()

	e = txn.Ctx.Set("req", req)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	c, e := txn.Do()
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	var vol volume.Volinfo
	e = c.Get("volinfo", &vol)
	if e == nil {
		rest.SendHTTPResponse(w, http.StatusCreated, vol)
		c.Logger().WithField("volname", vol.Name).Info("new volume created")
	} else {
		rest.SendHTTPError(w, http.StatusInternalServerError, "failed to get volinfo")
	}

	return
}
