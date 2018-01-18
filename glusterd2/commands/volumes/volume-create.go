package volumecommands

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
)


// VolCreateRequest defines the parameters for creating a volume in the volume-create command
type VolCreateRequest struct {
	Name         string            `json:"name"`
	Transport    string            `json:"transport,omitempty"`
	ReplicaCount int               `json:"replica,omitempty"`
	Bricks       []string          `json:"bricks,omitempty"`
	Force        bool              `json:"force,omitempty"`
	Options      map[string]string `json:"options,omitempty"`
	Size         int               `json:"size,omitempty"`
	// Bricks list is ordered (like in glusterd1) and decides which bricks
	// form replica sets.
}

func unmarshalVolCreateRequest(msg *api.VolCreateReq, r *http.Request) (int, error) {
	if err := restutils.UnmarshalRequest(r, msg); err != nil {
		return 422, gderrors.ErrJSONParsingFailed
	}

	if msg.Name == "" {
		return http.StatusBadRequest, gderrors.ErrEmptyVolName
	}

	if len(msg.Subvols) <= 0 {
		return http.StatusBadRequest, gderrors.ErrEmptyBrickList
	}

	for _, subvol := range msg.Subvols {
		if len(subvol.Bricks) <= 0 {
			return http.StatusBadRequest, gderrors.ErrEmptyBrickList
		}
	}
	return 0, nil

}

func voltypeFromSubvols(req *api.VolCreateReq) volume.VolType {
	if len(req.Subvols) == 0 {
		return volume.Distribute
	}
	numSubvols := len(req.Subvols)

	// TODO: Don't know how to decide on Volume Type if each subvol is different
	// For now just picking the first subvols Type, which satisfies
	// most of today's needs
	switch req.Subvols[0].Type {
	case "replicate":
		if numSubvols > 1 {
			return volume.DistReplicate
		}
		return volume.Replicate
	case "distribute":
		return volume.Distribute
	default:
		return volume.Distribute
	}
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

	v.DistCount = len(req.Subvols)

	v.Type = voltypeFromSubvols(req)

	for idx, subvolreq := range req.Subvols {
		if subvolreq.ReplicaCount == 0 && subvolreq.Type == "replicate" {
			return nil, errors.New("Replica count not specified")
		}

		if subvolreq.ReplicaCount > 0 && subvolreq.ReplicaCount != len(subvolreq.Bricks) {
			return nil, errors.New("Invalid number of bricks")
		}

		name := fmt.Sprintf("%s-%s-%d", v.Name, strings.ToLower(subvolreq.Type), idx)

		ty := volume.SubvolDistribute
		switch subvolreq.Type {
		case "replicate":
			ty = volume.SubvolReplicate
		case "disperse":
			ty = volume.SubvolDisperse
		default:
			ty = volume.SubvolDistribute
		}

		s := volume.Subvol{
			Name: name,
			ID:   uuid.NewRandom(),
			Type: ty,
		}

		if subvolreq.ArbiterCount != 0 {
			if subvolreq.ReplicaCount != 3 || subvolreq.ArbiterCount != 1 {
				return nil, errors.New("For arbiter configuration, replica count must be 3 and arbiter count must be 1. The 3rd brick of the replica will be the arbiter")
			}
			s.ArbiterCount = 1
		}

		if subvolreq.ReplicaCount == 0 {
			s.ReplicaCount = 1
		} else {
			s.ReplicaCount = subvolreq.ReplicaCount
		}

		if subvolreq.DisperseCount != 0 || subvolreq.DisperseData != 0 || subvolreq.DisperseRedundancy != 0 {
			err = checkDisperseParams(&subvolreq, &s)
			if err != nil {
				return nil, err
			}
		}
		s.Bricks, err = volume.NewBrickEntriesFunc(subvolreq.Bricks, v.Name, v.ID)
		if err != nil {
			return nil, err
		}
		v.Subvols = append(v.Subvols, s)

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
	if _, err = volume.ValidateBrickEntriesFunc(volinfo.GetBricks(), volinfo.ID, req.Force); err != nil {
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

	// TODO: Clean xattrs set if any. ValidateBrickEntriesFunc()
	// does a lot of things that it's not supposed to do.

	return nil
}

func registerVolCreateStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-create.Validate", validateVolumeCreate},
		{"vol-create.StoreVolume", storeVolume},
		{"vol-create.Rollback", rollBackVolumeCreate},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

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

	nodes, err := nodesFromVolumeCreateReq(req)
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

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-create.Validate",
			Nodes:  nodes,
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

	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("volume create transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err.Error(), api.ErrCodeDefault)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		}
		return
	}

	if err = txn.Ctx.Get("volinfo", &vol); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "failed to get volinfo", api.ErrCodeDefault)
		return
	}

	txn.Ctx.Logger().WithField("volname", vol.Name).Info("new volume created")
	events.Broadcast(newVolumeEvent(eventVolumeCreated, vol))

	resp := createVolumeCreateResp(vol)
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)
}

func createVolumeCreateResp(v *volume.Volinfo) *api.VolumeCreateResp {
	return (*api.VolumeCreateResp)(createVolumeInfoResp(v))
}

func checkDisperseParams(req *api.SubvolReq, s *volume.Subvol) error {
	count := len(req.Bricks)

	if req.DisperseData > 0 {
		if req.DisperseCount > 0 && req.DisperseRedundancy > 0 {
			if req.DisperseCount != req.DisperseData+req.DisperseRedundancy {
				return errors.New("Disperse count should be equal to sum of disperse-data and redundancy")
			}
		} else if req.DisperseRedundancy > 0 {
			req.DisperseCount = req.DisperseData + req.DisperseRedundancy
		} else if req.DisperseCount > 0 {
			req.DisperseRedundancy = req.DisperseCount - req.DisperseData
		} else {
			if count-req.DisperseData >= req.DisperseData {
				return errors.New("Need redundancy count along with disperse-data")
			}
			req.DisperseRedundancy = count - req.DisperseData
			req.DisperseCount = count
		}
	}

	if req.DisperseCount <= 0 {
		if count < 3 {
			return errors.New("Number of bricks must be greater than 2")
		}
		req.DisperseCount = count
	}

	if req.DisperseRedundancy <= 0 {
		req.DisperseRedundancy = int(getRedundancy(uint(req.DisperseCount)))
	}

	if req.DisperseCount != count {
		return errors.New("Disperse count and the number of bricks must be same for a pure disperse volume")
	}

	if 2*req.DisperseRedundancy >= req.DisperseCount {
		return errors.New("Invalid redundancy value")
	}

	s.DisperseCount = req.DisperseCount
	s.RedundancyCount = req.DisperseRedundancy

	return nil
}

func getRedundancy(disperse uint) uint {
	var temp, l, mask uint
	temp = disperse
	l = 0
	for temp = temp >> 1; temp != 0; temp = temp >> 1 {
		l = l + 1
	}
	mask = ^(1 << l)
	if red := disperse & mask; red != 0 {
		return red
	}
	return 1
}
