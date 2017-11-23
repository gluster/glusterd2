package volumecommands

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
)

func unmarshalVolCreateRequest(msg *api.VolCreateReq, r *http.Request) (int, error) {
	if err := restutils.UnmarshalRequest(r, msg); err != nil {
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

func createVolinfo(req *api.VolCreateReq) (*volume.Volinfo, error) {

	var err error

	v := new(volume.Volinfo)
	if req.Options != nil {
		v.Options = req.Options
	} else {
		v.Options = make(map[string]string)
	}
	v.ID = uuid.NewRandom()
	v.Name = req.Name

	if len(req.Transport) > 0 {
		v.Transport = req.Transport
	} else {
		v.Transport = "tcp"
	}

	if req.Replica == 0 {
		v.ReplicaCount = 1
	} else {
		v.ReplicaCount = req.Replica
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

	v.State = volume.VolCreated

	return v, nil
}

func validateVolumeCreate(c transaction.TxnCtx) error {

	var req api.VolCreateReq
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

func rollBackVolumeCreate(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	for _, b := range volinfo.Bricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}
		volgen.DeleteBrickVolfile(&b)
		// TODO: Clean xattrs set if any. ValidateBrickEntriesFunc()
		// does a lot of things that it's not supposed to do.
	}

	return nil
}

func registerVolCreateStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-create.Validate", validateVolumeCreate},
		{"vol-create.GenerateBrickVolfiles", generateBrickVolfiles},
		{"vol-create.StoreVolume", storeVolume},
		{"vol-create.Rollback", rollBackVolumeCreate},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := restutils.GetReqLogger(ctx)

	req := new(api.VolCreateReq)
	httpStatus, err := unmarshalVolCreateRequest(req, r)
	if err != nil {
		logger.WithError(err).Error("Failed to unmarshal volume request")
		restutils.SendHTTPError(ctx, w, httpStatus, err.Error(), api.ErrCodeDefault)
		return
	}

	if volume.ExistsFunc(req.Name) {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, gderrors.ErrVolExists.Error(), api.ErrCodeDefault)
		return
	}

	nodes, err := nodesFromBricks(req.Bricks)
	if err != nil {
		logger.WithError(err).Error("could not prepare node list")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err := validateOptions(req.Options); err != nil {
		logger.WithField("option", err.Error()).Error("invalid volume option specified")
		msg := fmt.Sprintf("invalid volume option specified: %s", err.Error())
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, msg, api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	lock, unlock, err := transaction.CreateLockSteps(req.Name)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = nodes
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-create.Validate",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc:   "vol-create.GenerateBrickVolfiles",
			UndoFunc: "vol-create.Rollback",
			Nodes:    txn.Nodes,
		},
		{
			DoFunc: "vol-create.StoreVolume",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}

	err = txn.Ctx.Set("req", req)
	if err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	vol, err := createVolinfo(req)
	if err != nil {
		logger.WithError(err).Error("failed to create volinfo")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	err = txn.Ctx.Set("volinfo", vol)
	if err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	c, err := txn.Do()
	if err != nil {
		logger.WithError(err).Error("volume create transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err.Error(), api.ErrCodeDefault)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		}
		return
	}

	if err = c.Get("volinfo", &vol); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "failed to get volinfo", api.ErrCodeDefault)
		return
	}

	c.Logger().WithField("volname", vol.Name).Info("new volume created")

	resp := createVolumeCreateResp(vol)
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)
}

func createVolumeCreateResp(v *volume.Volinfo) *api.VolumeCreateResp {
	return (*api.VolumeCreateResp)(createVolumeInfoResp(v))
}
