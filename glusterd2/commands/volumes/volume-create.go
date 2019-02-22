package volumecommands

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/bricksplanner"
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	transactionv2 "github.com/gluster/glusterd2/glusterd2/transactionv2"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	gutils "github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
	"go.opencensus.io/trace"
)

const (
	maxMetadataSizeLimit = 4 * gutils.KiB
	minVolumeSize        = 20 * gutils.MiB
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
		sf   oldtransaction.StepFunc
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
		oldtransaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	var err error

	var req api.VolCreateReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrJSONParsingFailed)
		return
	}

	if status, err := CreateVolume(ctx, req); err != nil {
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

// CreateVolume creates a volume
func CreateVolume(ctx context.Context, req api.VolCreateReq) (status int, err error) {
	ctx, span := trace.StartSpan(ctx, "/volumeCreateHandler")
	defer span.End()

	if err := validateVolCreateReq(&req); err != nil {
		return http.StatusBadRequest, err
	}

	if containsReservedGroupProfile(req.Options) {
		return http.StatusBadRequest, gderrors.ErrReservedGroupProfile
	}

	if req.ProvisionerType == "" {
		req.ProvisionerType = api.ProvisionerTypeLvm
	}

	if req.Size > 0 {
		applyDefaults(&req)

		if req.SnapshotReserveFactor < 1 {
			return http.StatusBadRequest, errors.New("invalid snapshot reserve factor")
		}

		if err := bricksplanner.PlanBricks(&req); err != nil {
			return http.StatusInternalServerError, err
		}
	} else {
		if err := checkDupBrickEntryVolCreate(req); err != nil {
			return http.StatusBadRequest, err
		}
	}

	req.Options, err = expandGroupOptions(req.Options)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if err := validateOptions(req.Options, req.VolOptionFlags); err != nil {
		return http.StatusBadRequest, err
	}

	// Include default Volume Options profile
	if len(req.Subvols) > 0 {
		groupProfile, exists := defaultGroupOptions["profile.default."+req.Subvols[0].Type]
		if exists {
			for _, opt := range groupProfile.Options {
				// Apply default option only if not overridden in volume create request
				_, exists = req.Options[opt.Name]
				if !exists {
					req.Options[opt.Name] = opt.OnValue
				}
			}
		}
	}

	nodes, err := req.Nodes()
	if err != nil {
		return http.StatusBadRequest, err
	}

	txn, err := transactionv2.NewTxnWithLocks(ctx, req.Name)
	if err != nil {
		return restutils.ErrToStatusCode(err)
	}
	defer txn.Done()

	if volume.Exists(req.Name) {
		return http.StatusBadRequest, gderrors.ErrVolExists
	}

	txn.Steps = []*oldtransaction.Step{
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
			// Need to wait for volinfo to be created first
			Sync: true,
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
			Sync:     true,
		},
	}

	if err := txn.Ctx.Set("req", &req); err != nil {
		return http.StatusInternalServerError, err
	}

	// Add attributes to the span with info that can be viewed along with traces.
	// The attributes can also be used to filter traces on the tracing UI.
	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
		trace.StringAttribute("volName", req.Name),
	)

	if err := txn.Do(); err != nil {
		return restutils.ErrToStatusCode(err)
	}

	return http.StatusCreated, nil
}
