package volumecommands

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	customerror "github.com/gluster/glusterd2/pkg/errors"
	rebalance "github.com/gluster/glusterd2/plugins/rebalance"
	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func registerVolShrinkStepFuncs() {
	transaction.RegisterStepFunc(storeVolume, "vol-shrink.UpdateVolinfo")
	transaction.RegisterStepFunc(notifyVolfileChange, "vol-shrink.NotifyClients")
	transaction.RegisterStepFunc(startRebalance, "vol-shrink.StartRebalance")
}

func startRebalance(c transaction.TxnCtx) error {
	var rinfo rebalanceapi.RebalInfo
	err := c.Get("rinfo", &rinfo)
	if err != nil {
		return err
	}

	rebalanceProcess, err := rebalance.NewRebalanceProcess(rinfo)
	if err != nil {
		return err
	}

	err = daemon.Start(rebalanceProcess, true)

	return err
}

func validateVolumeShrinkReq(req api.VolShrinkReq) error {
	dupEntry := map[string]bool{}

	for _, brick := range req.Bricks {
		if dupEntry[brick.PeerID+filepath.Clean(brick.Path)] == true {
			return customerror.ErrDuplicateBrickPath
		}
		dupEntry[brick.PeerID+filepath.Clean(brick.Path)] = true

	}

	return nil

}

func volumeShrinkHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	volname := mux.Vars(r)["volname"]

	var req api.VolShrinkReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := validateVolumeShrinkReq(req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
		return
	}

	for index := range req.Bricks {
		for _, b := range req.Bricks {
			isPresent := false
			for _, brick := range volinfo.Subvols[index].Bricks {
				if brick.PeerID.String() == b.PeerID && brick.Path == filepath.Clean(b.Path) {
					isPresent = true
					break
				}
			}
			if !isPresent {
				restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "One or more brick is not part of given volume")
				return
			}
		}
	}

	switch volinfo.Type {
	case volume.Distribute:
	case volume.Replicate:
	case volume.DistReplicate:
		if len(req.Bricks)%volinfo.Subvols[0].ReplicaCount != 0 {
			err := errors.New("wrong number of bricks to remove")
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
			return
		}
	default:
		err := errors.New("not implemented: " + volinfo.Type.String())
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return

	}

	nodes, err := req.Nodes()
	if err != nil {
		logger.WithError(err).Error("could not prepare node list")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-shrink.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-shrink.NotifyClients",
			Nodes:  nodes,
		},
		{
			DoFunc: "vol-shrink.StartRebalance",
			Nodes:  nodes,
		},
	}

	decommissionedSubvols, err := findDecommissioned(req.Bricks, volinfo)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return

	}

	// The following line is for testing purposes.
	// It seems that there is no other way to include this information in the rebalance volfile right now.
	volinfo.Options["distribute.decommissioned-bricks"] = strings.TrimSpace(decommissionedSubvols)

	var rinfo rebalanceapi.RebalInfo
	rinfo.Volname = volname
	rinfo.RebalanceID = uuid.NewRandom()
	rinfo.Cmd = rebalanceapi.CmdStartForce
	rinfo.State = rebalanceapi.NotStarted
	rinfo.CommitHash = rebalance.SetCommitHash()
	if err := txn.Ctx.Set("rinfo", rinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).Error("remove bricks start transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, decommissionedSubvols)

}

func findDecommissioned(bricks []api.BrickReq, volinfo *volume.Volinfo) (string, error) {
	brickSet := make(map[string]bool)
	for _, brick := range bricks {
		u := uuid.Parse(brick.PeerID)
		if u == nil {
			return "", errors.New("Invalid nodeid")
		}
		path, err := filepath.Abs(brick.Path)
		if err != nil {
			return "", err
		}
		brickSet[brick.PeerID+":"+path] = true
	}

	var subvolMap = make(map[string]int)
	for _, subvol := range volinfo.Subvols {
		for _, b := range subvol.Bricks {
			if brickSet[b.PeerID.String()+":"+b.Path] {
				if count, ok := subvolMap[subvol.Name]; !ok {
					subvolMap[subvol.Name] = 1
				} else {
					subvolMap[subvol.Name] = count + 1
				}
			}

		}
	}

	var base int
	switch volinfo.Type {
	case volume.Distribute:
		base = 1
	case volume.Replicate:
		base = len(bricks)
	case volume.DistReplicate:
		base = volinfo.Subvols[0].ReplicaCount
	default:
		return "", errors.New("not implemented: " + volinfo.Type.String())
	}

	decommissioned := ""
	for subvol, count := range subvolMap {
		if count != base {
			return "", errors.New("Wrong number of bricks in the subvolume")
		}
		decommissioned = decommissioned + subvol + " "
	}

	return decommissioned, nil
}
