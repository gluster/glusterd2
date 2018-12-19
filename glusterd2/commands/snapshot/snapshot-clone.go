package snapshotcommands

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/lvmutils"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

func undoSnapshotClone(c transaction.TxnCtx) error {
	var volinfo volume.Volinfo

	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {
		volume.UmountBrick(b)
		if err := lvmutils.RemoveLVSnapshot(b.MountInfo.DevicePath); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Debug("Failed to remove snapshotted LVM")
			return err
		}
	}

	return nil
}
func undoStoreSnapshotClone(c transaction.TxnCtx) error {
	var (
		vol volume.Volinfo
		err error
	)

	if err = c.Get("volinfo", &vol); err != nil {
		return err
	}

	err = volume.DeleteVolume(vol.Name)
	return err
}

func storeSnapshotClone(c transaction.TxnCtx) error {
	var vol volume.Volinfo
	if err := c.Get("volinfo", &vol); err != nil {
		return err
	}
	if err := volume.AddOrUpdateVolumeFunc(&vol); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", vol.Name).Error("storeVolume: failed to store Volinfo")
		return err
	}

	return nil
}
func takeSnapshotClone(c transaction.TxnCtx) error {
	var snapname string
	var newVol volume.Volinfo

	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}

	if err := c.Get("volinfo", &newVol); err != nil {
		return err
	}

	snapVol := &snapinfo.SnapVolinfo
	if err = takeVolumeSnapshots(&newVol, snapVol); err != nil {
		return err
	}

	err = volume.MountVolumeBricks(&newVol, false)
	return err
}

func populateCloneBrickMountData(volinfo *volume.Volinfo, name string) (map[string]snapshot.BrickMountData, error) {
	nodeData := make(map[string]snapshot.BrickMountData)

	for svIdx, sv := range volinfo.Subvols {
		for bIdx, b := range sv.Bricks {
			if !uuid.Equal(b.PeerID, gdctx.MyUUID) {
				continue
			}

			mountRoot, err := volume.GetBrickMountRoot(b.Path)
			if err != nil {
				return nil, err
			}
			brickDirSuffix := b.Path[len(mountRoot):]
			mntInfo, err := volume.GetBrickMountInfo(mountRoot)
			if err != nil {
				log.WithError(err).WithField(
					"brick", b.Path,
				).Error("Failed to get mount information")

				return nil, err
			}

			suffix := fmt.Sprintf("clone_%s_%s_s%d_b%d", name, volinfo.Name, svIdx+1, bIdx+1)

			devicePath, err := lvmutils.CreateDevicePath(mntInfo.FsName, suffix)
			if err != nil {
				log.WithError(err).WithField(
					"deviceName", devicePath,
				).Error("Failed to create device name. A thinLV with same name exist")
				return nil, err
			}
			nodeData[b.String()] = snapshot.BrickMountData{
				BrickDirSuffix: brickDirSuffix,
				DevicePath:     devicePath,
				FsType:         mntInfo.MntType,
				MntOpts:        updateMntOps(mntInfo.MntType, mntInfo.MntOpts),
				Path:           snapshotCloneBrickCreate(name, brickDirSuffix, svIdx+1, bIdx+1),
			}
		}
	}
	return nodeData, nil
}

func validateSnapClone(c transaction.TxnCtx) error {
	var (
		statusStr           []string
		err                 error
		snapname, clonename string
		nodeData            map[string]snapshot.BrickMountData
	)

	if err = lvmutils.CommonPrevalidation(lvmutils.CreateCommand); err != nil {
		log.WithError(err).WithField(
			"command", lvmutils.CreateCommand,
		).Error("Failed to find lvm packages")
		return err
	}

	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	if err := c.Get("clonename", &clonename); err != nil {
		return err
	}

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}
	volinfo := &snapinfo.SnapVolinfo

	brickStatuses, err := volume.CheckBricksStatus(volinfo)
	if err != nil {
		return err
	}

	for _, brickStatus := range brickStatuses {
		if brickStatus.Online == false {
			statusStr = append(statusStr, brickStatus.Info.String())
		}
	}
	if statusStr != nil {
		log.WithError(err).WithField(
			"Bricks", statusStr,
		).Error("Bricks are offline")

		return errors.New("one or more brick is offline")
	}

	if nodeData, err = populateCloneBrickMountData(volinfo, clonename); err != nil {
		return err
	}
	c.SetNodeResult(gdctx.MyUUID, snapshot.NodeDataTxnKey, &nodeData)
	//TODO Quorum check has to be implemented once we implement highly available snapshot
	return nil
}

func createCloneVolinfo(c transaction.TxnCtx) error {
	var clonename, snapname string
	nodeData := make(map[string]snapshot.BrickMountData)

	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	if err := c.Get("clonename", &clonename); err != nil {
		return err
	}
	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}
	volinfo := &snapinfo.SnapVolinfo

	for _, node := range volinfo.Nodes() {
		tmp := make(map[string]snapshot.BrickMountData)
		if err := c.GetNodeResult(node, snapshot.NodeDataTxnKey, &tmp); err != nil {
			return err
		}
		for k, v := range tmp {
			nodeData[k] = v
		}
	}

	newVol := new(volume.Volinfo)
	duplicateVolinfo(volinfo, newVol)

	for key, value := range snapinfo.OptionChange {
		newVol.Options[key] = value
	}

	newVol.State = volume.VolCreated
	newVol.GraphMap = volinfo.GraphMap
	newVol.ID = uuid.NewRandom()
	newVol.Name = clonename
	newVol.VolfileID = clonename

	if err = createSnapSubvols(newVol, volinfo, nodeData); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"snapshot":    snapname,
			"volume name": clonename,
		}).Error("Failed to create clone volinfo")

		return err
	}

	err = c.Set("volinfo", newVol)
	return err
}

func snapshotCloneBrickCreate(cloneName, brickDirSuffix string, subvolNumber, brickNumber int) string {
	cloneDirPrefix := config.GetString("rundir") + "/clones"
	brickPath := fmt.Sprintf("%s/%s/subvol%d/brick%d%s", cloneDirPrefix, cloneName, subvolNumber, brickNumber, brickDirSuffix)
	return brickPath
}

func registerSnapCloneStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"snap-clone.Validate", validateSnapClone},
		{"snap-clone.CreateCloneVolinfo", createCloneVolinfo},
		{"snap-clone.TakeBrickSnapshots", takeSnapshotClone},
		{"snap-clone.UndoBrickSnapshots", undoSnapshotClone},
		{"snap-clone.StoreSnapshot", storeSnapshotClone},
		{"snap-clone.UndoStoreSnapshotOnClone", undoStoreSnapshotClone},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func snapshotCloneHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := new(api.SnapCloneReq)
	logger := gdctx.GetReqLogger(ctx)

	snapname := mux.Vars(r)["snapname"]
	if snapname == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.New("snapshot name should not be empty"))
		return
	}

	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrJSONParsingFailed)
		return
	}

	if !volume.IsValidName(req.CloneName) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrInvalidVolName)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, req.CloneName, snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	snapVol := &snapinfo.SnapVolinfo

	if volume.Exists(req.CloneName) {
		errMsg := "A volume with the same clone name exist."
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
		return
	}

	if snapVol.State != volume.VolStarted {
		errMsg := "Snapshot must be in started state before cloning."
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
		return
	}
	txn.Nodes = snapVol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "snap-clone.Validate",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "snap-clone.CreateCloneVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
			Sync:   true,
		},
		{
			DoFunc:   "snap-clone.TakeBrickSnapshots",
			UndoFunc: "snap-clone.UndoBrickSnapshots",
			Nodes:    txn.Nodes,
		},
		{
			DoFunc:   "snap-clone.StoreSnapshot",
			UndoFunc: "snap-clone.UndoStoreSnapshotOnClone",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
			Sync:     true,
		},
	}
	if err = txn.Ctx.Set("snapname", &snapname); err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if err = txn.Ctx.Set("clonename", &req.CloneName); err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).Error("snapshot clone transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	txn.Ctx.Logger().WithField("CloneName", req.CloneName).Info("new volume cloned from snapshot")

	vol, err := volume.GetVolume(req.CloneName)
	if err != nil {
		// FIXME: If volume was created successfully in the txn above and
		// then the store goes down by the time we reach here, what do
		// we return to the client ?
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	resp := createSnapshotCloneResp(vol)
	restutils.SetLocationHeader(r, w, vol.Name)
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)

}

func createSnapshotCloneResp(v *volume.Volinfo) *api.SnapshotCloneResp {
	return (*api.SnapshotCloneResp)(volume.CreateVolumeInfoResp(v))
}
