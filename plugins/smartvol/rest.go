package smartvol

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	smartvolapi "github.com/gluster/glusterd2/plugins/smartvol/api"
	"github.com/gluster/glusterd2/plugins/smartvol/bricksplanner"

	"github.com/pborman/uuid"
)

const (
	minVolumeSize = 10
)

func validateVolCreateReq(req *smartvolapi.VolCreateReq) error {
	if !volume.IsValidName(req.Name) {
		return gderrors.ErrInvalidVolName
	}

	if req.Transport != "" && req.Transport != "tcp" && req.Transport != "rdma" {
		return errors.New("invalid transport. Supported values: tcp or rdma")
	}

	if req.Size < minVolumeSize {
		return errors.New("invalid Volume Size, Minimum size required is " + strconv.Itoa(minVolumeSize))
	}

	return nil
}

func getVolumeReq(req *smartvolapi.Volume) api.VolCreateReq {
	reqVolCreate := api.VolCreateReq{
		Name:      req.Name,
		Transport: req.Transport,
		Subvols:   make([]api.SubvolReq, len(req.Subvols)),
		Force:     req.Force,
	}

	for sidx, sv := range req.Subvols {
		reqVolCreate.Subvols[sidx] = api.SubvolReq{
			Type:               sv.Type,
			Bricks:             make([]api.BrickReq, len(sv.Bricks)),
			ReplicaCount:       sv.ReplicaCount,
			ArbiterCount:       sv.ArbiterCount,
			DisperseCount:      sv.DisperseCount,
			DisperseData:       sv.DisperseDataCount,
			DisperseRedundancy: sv.DisperseRedundancyCount,
		}

		for bidx, b := range sv.Bricks {
			reqVolCreate.Subvols[sidx].Bricks[bidx] = api.BrickReq{
				Type:   b.Type,
				PeerID: b.PeerID,
				Path:   b.Path,
			}
		}
	}

	return reqVolCreate
}

func applyDefaults(req *smartvolapi.VolCreateReq) {
	if req.SnapshotReserveFactor == 0 {
		req.SnapshotReserveFactor = 1
	}

	// Snapshot reserve not required if not enabled
	if !req.SnapshotEnabled {
		req.SnapshotReserveFactor = 1
	}
}

func smartVolumeCreateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	var err error

	var req smartvolapi.VolCreateReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := validateVolCreateReq(&req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err = volume.GetVolume(req.Name)
	if err == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrVolExists)
		return
	} else if err != gderrors.ErrVolNotFound {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	applyDefaults(&req)

	if req.SnapshotReserveFactor < 1 {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.New("invalid Snapshot Reserve Factor"))
		return
	}

	// Convert request to internal Structure
	smartvolReq := smartvolapi.Volume{
		Name:                    req.Name,
		Transport:               req.Transport,
		Force:                   req.Force,
		Size:                    req.Size,
		DistributeCount:         req.DistributeCount,
		ReplicaCount:            req.ReplicaCount,
		ArbiterCount:            req.ArbiterCount,
		DisperseCount:           req.DisperseCount,
		DisperseDataCount:       req.DisperseDataCount,
		DisperseRedundancyCount: req.DisperseRedundancyCount,
		SnapshotEnabled:         req.SnapshotEnabled,
		SnapshotReserveFactor:   req.SnapshotReserveFactor,
		LimitPeers:              req.LimitPeers,
		LimitZones:              req.LimitZones,
		ExcludePeers:            req.ExcludePeers,
		ExcludeZones:            req.ExcludeZones,
		SubvolZonesOverlap:      req.SubvolZonesOverlap,
	}

	if err := bricksplanner.PlanBricks(&smartvolReq); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	// Convert struct to traditional Volume Create format
	reqVolCreate := getVolumeReq(&smartvolReq)

	nodes, err := reqVolCreate.Nodes()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, req.Name)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	txn.Nodes = nodes
	txn.Steps = []*transaction.Step{
		{
			DoFunc:   "vol-create.PrepareBricks",
			UndoFunc: "vol-create.UndoPrepareBricks",
			Nodes:    nodes,
		},
		{
			DoFunc: "vol-create.CreateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-create.ValidateBricks",
			Nodes:  nodes,
		},
		{
			DoFunc:   "vol-create.InitBricks",
			UndoFunc: "vol-create.UndoInitBricks",
			Nodes:    nodes,
		},
		{
			DoFunc:   "vol-create.StoreVolume",
			UndoFunc: "vol-create.UndoStoreVolume",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
	}

	if err := txn.Ctx.Set("req", &reqVolCreate); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("reqBricksPrepare", &smartvolReq); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	volinfo, err := volume.GetVolume(req.Name)
	if err != nil {
		// FIXME: If volume was created successfully in the txn above and
		// then the store goes down by the time we reach here, what do
		// we return to the client ?
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("volume-name", volinfo.Name).Info("new volume created")

	events.Broadcast(volume.NewEvent(volume.EventVolumeCreated, volinfo))

	resp := (*api.VolumeCreateResp)(volume.CreateVolumeInfoResp(volinfo))
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)
}
