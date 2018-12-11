package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/bricksplanner"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func registerReplaceBrickStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"brick-replace.PrepareBricks", prepareBricks},
		{"brick-replace.ReplaceVolinfo", replaceVolinfo},
		{"brick-replace.StartBrick", startBrick},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func replaceBrickHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]

	if !volume.IsValidName(volname) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrInvalidVolName)
		return
	}

	// Unmarshal ReplaceBrickReq
	var req api.ReplaceBrickReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrJSONParsingFailed)
		return
	}

	if uuid.Parse(req.SrcPeerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "invalid peerID passed in url")
		return
	}

	if err := validateVolumeFlags(req.Flags); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	// Get Volume Info
	vol, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	subVols := vol.Subvols

	var srcBrickInfo brick.Brickinfo
	subVolIndex := 0
	brickIndex := 0
LOOP:
	for index := range subVols {
		for i, brick := range subVols[index].Bricks {
			// Get Source Brick Info
			if brick.PeerID.String() == req.SrcPeerID && brick.Path == req.SrcBrickPath {
				subVolIndex = index
				brickIndex = i
				srcBrickInfo = brick
				break LOOP
			}
		}
	}

	excludeZones := make([]string, 0)
	for svIndex, sv := range subVols {
		// if SubvolZonesOverlap is true then bricks of only that particular
		// subvolume will be considered.
		if req.SubvolZonesOverlap && subVolIndex != svIndex {
			continue
		}
		for _, b := range sv.Bricks {
			p, err := peer.GetPeer(b.PeerID.String())
			if err != nil {
				restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
				return
			}
			excludeZones = append(excludeZones, p.Metadata["_zone"])
		}
	}
	req.ExcludeZones = append(req.ExcludeZones, excludeZones...)

	subvolumes := make([]api.SubvolReq, 0)
	volreq := api.VolCreateReq{
		Subvols:      subvolumes,
		Size:         vol.Capacity,
		LimitPeers:   req.LimitPeers,
		LimitZones:   req.LimitZones,
		ExcludePeers: req.ExcludePeers,
		ExcludeZones: req.ExcludeZones,
	}
	availableVgs, err := bricksplanner.GetAvailableVgs(&volreq)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	// TODO: check for available vgs in zones already being used in volume.
	if len(availableVgs) == 0 {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "No volume groups are available")
		return
	}

	mtabEntries, err := volume.GetMounts()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	// Get source brick information like size etc
	brickInfo, err := volume.BrickStatus(srcBrickInfo, mtabEntries)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	// Get new brick from the available vgs
	newBrick := bricksplanner.GetNewBrick(availableVgs, brickInfo, vol, subVolIndex, brickIndex)

	peerID := uuid.Parse(newBrick.PeerID)
	if peerID == nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "peer id of new brick could not be parsed")
		return
	}
	allPeerIDs := vol.Nodes()
	nodes := []uuid.UUID{peerID}
	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "brick-replace.PrepareBricks",
			Nodes:  nodes,
		},
		{
			DoFunc: "brick-replace.ReplaceVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc:   "vol-create.InitBricks",
			UndoFunc: "vol-create.UndoInitBricks",
			Nodes:    nodes,
		},
		{
			DoFunc: "vol-start.StartBricks",
			Nodes:  nodes,
		},
		{
			DoFunc:   "vol-create.StoreVolume",
			UndoFunc: "vol-create.UndoStoreVolume",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-expand.NotifyClients",
			Nodes:  allPeerIDs,
		},
	}

	if err = txn.Ctx.Set("newBrick", &newBrick); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if err = txn.Ctx.Set("srcBrickInfo", &srcBrickInfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if err = txn.Ctx.Set("subVolIndex", &subVolIndex); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if err = txn.Ctx.Set("brickIndex", &brickIndex); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if err = txn.Ctx.Set("volinfo", &vol); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).WithField("volume-name", volname).Error("replace brick transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	resp := createReplaceBrickResp(vol)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)

	return
}

// Replace brick resp
func createReplaceBrickResp(v *volume.Volinfo) *api.ReplaceBrickResp {
	return (*api.ReplaceBrickResp)(volume.CreateVolumeInfoResp(v))
}
