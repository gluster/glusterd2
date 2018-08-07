package volumecommands

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/bricksplanner"
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
	"go.opencensus.io/trace"
)

const (
	maxMetadataSizeLimit = 4096
	minVolumeSize        = 20
)

func applyDefaults(req *api.VolCreateReq) {
	if req.SnapshotReserveFactor == 0 {
		req.SnapshotReserveFactor = 1
	}

	// Snapshot reserve not required if not enabled
	if !req.SnapshotEnabled {
		req.SnapshotReserveFactor = 1
	}
}

func validateVolCreateReq(req *api.VolCreateReq) error {
	if !volume.IsValidName(req.Name) {
		return gderrors.ErrInvalidVolName
	}

	if req.Transport != "" && req.Transport != "tcp" && req.Transport != "rdma" {
		return errors.New("invalid transport. Supported values: tcp or rdma")
	}

	if req.Size > 0 && req.Size < minVolumeSize {
		return errors.New("invalid Volume Size, Minimum size required is " + strconv.Itoa(minVolumeSize))
	}

	if req.Size == 0 && len(req.Subvols) <= 0 {
		return gderrors.ErrEmptyBrickList
	}

	for _, subvol := range req.Subvols {
		if len(subvol.Bricks) <= 0 {
			return gderrors.ErrEmptyBrickList
		}
	}
	if req.MetadataSize() > maxMetadataSizeLimit {
		return gderrors.ErrMetadataSizeOutOfBounds
	}

	return validateVolumeFlags(req.Flags)
}

func checkDupBrickEntryVolCreate(req api.VolCreateReq) error {
	dupEntry := map[string]bool{}

	for index := range req.Subvols {
		for _, brick := range req.Subvols[index].Bricks {
			if dupEntry[brick.PeerID+filepath.Clean(brick.Path)] == true {
				return gderrors.ErrDuplicateBrickPath
			}
			dupEntry[brick.PeerID+filepath.Clean(brick.Path)] = true

		}
	}
	return nil
}
func registerVolCreateStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-create.CreateVolinfo", createVolinfo},
		{"vol-create.ValidateBricks", validateBricks},
		{"vol-create.InitBricks", initBricks},
		{"vol-create.UndoInitBricks", undoInitBricks},
		{"vol-create.StoreVolume", storeVolume},
		{"vol-create.UndoStoreVolume", undoStoreVolumeOnCreate},
		{"vol-create.PrepareBricks", txnPrepareBricks},
		{"vol-create.UndoPrepareBricks", txnUndoPrepareBricks},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, span := trace.StartSpan(ctx, "/volumeCreateHandler")
	defer span.End()

	logger := gdctx.GetReqLogger(ctx)
	var err error

	var req api.VolCreateReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrJSONParsingFailed)
		return
	}

	// Generate Volume name if not provided
	if req.Name == "" {
		req.Name = volume.GenerateVolumeName()
	}

	if err := validateVolCreateReq(&req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	if req.Size > 0 {
		applyDefaults(&req)

		if req.SnapshotReserveFactor < 1 {
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest,
				errors.New("invalid snapshot reserve factor"))
			return
		}

		if err := bricksplanner.PlanBricks(&req); err != nil {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
			return
		}
	} else {
		if err := checkDupBrickEntryVolCreate(req); err != nil {
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
			return
		}
	}

	req.Options, err = expandGroupOptions(req.Options)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := validateOptions(req.Options, req.Advanced, req.Experimental, req.Deprecated); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	nodes, err := req.Nodes()
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

	txn.Steps = []*transaction.Step{
		{
			DoFunc:   "vol-create.PrepareBricks",
			UndoFunc: "vol-create.UndoPrepareBricks",
			Nodes:    nodes,
			Skip:     (req.Size == 0),
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

	if err := txn.Ctx.Set("req", &req); err != nil {
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

	resp := createVolumeCreateResp(volinfo)
	restutils.SetLocationHeader(r, w, volinfo.Name)
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)
}

func createVolumeCreateResp(v *volume.Volinfo) *api.VolumeCreateResp {
	return (*api.VolumeCreateResp)(volume.CreateVolumeInfoResp(v))
}
