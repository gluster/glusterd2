package volumecommands

import (
	"errors"
	"net/http"

	"github.com/gluster/glusterd2/context"
	gderrors "github.com/gluster/glusterd2/errors"
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

func validateVolumeCreate(c *context.Context) error {
	req, ok := c.Get("req").(*volume.VolCreateRequest)
	if !ok {
		return errors.New("failed to get request from context")
	}

	if volume.ExistsFunc(req.Name) {
		c.Log.WithField("volume", req.Name).Error("volume already exists")
		return gderrors.ErrVolExists
	}

	vol, err := createVolinfo(req)
	if err != nil {
		return err
	}

	_, err = volume.ValidateBrickEntriesFunc(vol.Bricks, vol.ID, req.Force)
	if err != nil {
		return err
	}

	// Store volinfo for later usage
	c.Set("volinfo", vol)

	return nil
}

func generateVolfiles(c *context.Context) error {
	vol, ok := c.Get("volinfo").(*volume.Volinfo)
	if !ok {
		return errors.New("failed to get volinfo from context")
	}

	// Creating client and server volfile
	e := volgen.GenerateVolfileFunc(vol)
	if e != nil {
		c.Log.WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("failed to generate volfile")
		return e
	}
	return nil
}

func storeVolume(c *context.Context) error {
	vol, ok := c.Get("volinfo").(*volume.Volinfo)
	if !ok {
		return errors.New("failed to get volinfo from context")
	}

	e := volume.AddOrUpdateVolumeFunc(vol)
	if e != nil {
		c.Log.WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("Failed to create volume")
		return e
	}

	log.WithField("volume", vol.Name).Debug("new volume added")
	return nil
}

func rollBackVolumeCreate(c *context.Context) error {
	vol, ok := c.Get("volinfo").(*volume.Volinfo)
	if !ok {
		return errors.New("failed to get volinfo from context")
	}

	_ = volume.RemoveBrickPaths(vol.Bricks)

	return nil
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {
	req := new(volume.VolCreateRequest)

	httpStatus, e := unmarshalVolCreateRequest(req, r)
	if e != nil {
		rest.SendHTTPError(w, httpStatus, e.Error())
		return
	}

	// TODO: Properly construct these things
	nodes := make([]string, 1)
	c := context.NewLoggingContext(log.Fields{
		"reqid": uuid.NewRandom().String(),
	})
	c.Set("req", req)
	c.Set("lockKey", req.Name)
	txn := transaction.GetTxn("VolumeCreate")
	if txn == nil {
		c.Log.Error("Failed to get transaction object for volumeCreate")
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	// Update with context
	txn.UpdateTxn(c, nodes)

	if e := txn.Do(); e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	vol, ok := c.Get("volinfo").(*volume.Volinfo)
	if ok {
		rest.SendHTTPResponse(w, http.StatusCreated, vol)
		c.Log.WithField("volname", vol.Name).Info("new volume created")
	} else {
		rest.SendHTTPError(w, http.StatusInternalServerError, "failed to get volinfo")
	}

	return
}
