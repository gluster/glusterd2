package snapshotcommands

/*
TODO
*setactiveonskip flag
*snap max limit
*snap soft limit
*snap auto-delete
*activate-on-create
*/

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc/dict"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/snapshot/lvm"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

func barrierActivateDeactivateFunc(volinfo *volume.Volinfo, option string, originUUID uuid.UUID) error {
	var req brick.GfBrickOpReq
	var err error

	if bytes.Equal(originUUID, gdctx.MyUUID) {
		volinfo.Options["features.barrier"] = option
		if err = volume.AddOrUpdateVolumeFunc(volinfo); err != nil {
			log.WithError(err).WithField(
				"volume", volinfo.Name).Debug("failed to store volume info")
			return err
		}

		if err = volgen.Generate(); err != nil {
			log.WithError(err).WithField(
				"volume", volinfo.Name).Debug("failed to generate volfiles")
			return err
		}
	}

	reqDict := make(map[string]string)
	reqDict["barrier"] = option
	req.Op = int(brick.OpBrickBarrier)
	req.Input, err = dict.Serialize(reqDict)
	if err != nil {
		log.WithError(err).WithField(
			"volume", volinfo.Name).Error("failed to serialize dict for barrier option")
	}

	for _, b := range volinfo.GetLocalBricks() {
		brickDaemon, err := brick.NewGlusterfsd(b)
		if err != nil {
			return err
		}

		client, err := daemon.GetRPCClient(brickDaemon)
		if err != nil {
			log.WithError(err).WithField(
				"brick", b.String()).Error("failed to connect to brick, Aborting the barrier config operation")
			return err
		}

		req.Name = b.Path

		var rsp brick.GfBrickOpRsp
		err = client.Call("Brick.OpBrickBarrier", req, &rsp)
		if err != nil || rsp.OpRet != 0 {
			log.WithError(err).WithField(
				"brick", b.String()).Error("failed to send barrier RPC")
			return err
		}

	}
	return nil
}
func deactivateBarrier(c transaction.TxnCtx) error {
	var barrierOp string
	var snapInfo snapshot.Snapinfo
	if err := c.Get("barrier-enabled", &barrierOp); err != nil {
		return err
	}

	if barrierOp == "enable" {
		/*
			Barrier is already enabled, Just return success
		*/
		return nil
	}
	/*
		Do we need to do this ?
	*/
	var originatorUUID uuid.UUID
	if err := c.Get("originator-uuid", &originatorUUID); err != nil {
		return err
	}

	if err := c.Get("snapinfo", &snapInfo); err != nil {
		return err
	}

	volinfo, err := volume.GetVolume(snapInfo.ParentVolume)
	if err != nil {
		return err
	}
	c.Logger().WithFields(log.Fields{"volume": volinfo.Name}).Info("Sending Barrier request to bricks")

	err = barrierActivateDeactivateFunc(volinfo, "disable", originatorUUID)
	if err != nil {
		return err
	}

	return nil

}

func activateBarrier(c transaction.TxnCtx) error {
	var barrierOp string
	var snapInfo snapshot.Snapinfo
	if err := c.Get("barrier-enabled", &barrierOp); err != nil {
		return err
	}

	if barrierOp == "enabled" {
		/*
			Barrier is already enabled, Just return success
		*/
		return nil
	}
	/*
		Do we need to do this ?
	*/
	var originatorUUID uuid.UUID
	if err := c.Get("originator-uuid", &originatorUUID); err != nil {
		return err
	}

	if err := c.Get("snapinfo", &snapInfo); err != nil {
		return err
	}

	volinfo, err := volume.GetVolume(snapInfo.ParentVolume)
	if err != nil {
		return err
	}
	c.Logger().WithFields(log.Fields{"volume": volinfo.Name}).Info("Sending Barrier request to bricks")

	err = barrierActivateDeactivateFunc(volinfo, "enable", originatorUUID)
	if err != nil {
		return err
	}

	return nil

}
func undoBrickSnapshots(c transaction.TxnCtx) error {
	var snapInfo snapshot.Snapinfo

	if err := c.Get("snapinfo", &snapInfo); err != nil {
		return err
	}

	snapVol := snapInfo.SnapVolinfo
	for _, b := range snapVol.GetLocalBricks() {
		if err := lvm.RemoveBrickSnapshot(b.MountInfo.DevicePath); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Debug("Failed to remove snapshotted LVM")
			return err
		}
	}

	return nil
}
func undoStoreSnapshotOnCreate(c transaction.TxnCtx) error {

	var snapInfo snapshot.Snapinfo
	if err := c.Get("snapinfo", &snapInfo); err != nil {
		return err
	}

	if err := snapshot.DeleteSnapshot(&snapInfo); err != nil {

		c.Logger().WithError(err).WithField(
			"snapshot", snapshot.GetStorePath(&snapInfo),
		).Warn("Failed to delete snapinfo from store")
		return err
	}

	if err := volgen.DeleteVolfiles(snapInfo.SnapVolinfo.VolfileID); err != nil {
		c.Logger().WithError(err).
			WithField("snapshot", snapshot.GetStorePath(&snapInfo)).
			Warn("failed to delete volfiles of snapshot")
	}

	return nil
}

// storeSnapshot uses to store the volinfo and to generate client volfile
func storeSnapshot(c transaction.TxnCtx) error {

	var snapInfo snapshot.Snapinfo
	if err := c.Get("snapinfo", &snapInfo); err != nil {
		return err
	}
	volinfo := &snapInfo.SnapVolinfo

	vol, err := volume.GetVolume(snapInfo.ParentVolume)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", snapInfo.ParentVolume).Debug("storeVolume: failed to fetch Volinfo from store")
		return err
	}

	vol.SnapList = append(vol.SnapList, volinfo.Name)
	if err := volume.AddOrUpdateVolumeFunc(vol); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", vol.Name).Debug("storeVolume: failed to store Volinfo")
		return err
	}

	if err := snapshot.AddOrUpdateSnapFunc(&snapInfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("storeSnapshot: failed to store snapshot info")
		return err
	}
	if err := volgen.Generate(); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("generateVolfiles: failed to generate volfiles")
		return err
	}

	return nil
}

func unmarshalSnapCreateRequest(msg *api.SnapCreateReq, r *http.Request) (int, error) {
	if err := restutils.UnmarshalRequest(r, msg); err != nil {
		return 422, gderrors.ErrJSONParsingFailed
	}

	if msg.VolName == "" {
		return http.StatusBadRequest, gderrors.ErrEmptyVolName
	}
	if msg.SnapName == "" {
		return http.StatusBadRequest, gderrors.ErrEmptySnapName
	}
	return 0, nil
}
func updateMntOps(FsType, MntOpts string) string {
	switch FsType {
	case "xfs":
		if len(MntOpts) > 0 {
			return (MntOpts + ",nouuid")
		}
		return "nouuid"

	case "ext4":
		fallthrough
	case "ext3":
		fallthrough
	case "ext2":
	default:
	}
	return MntOpts
}
func populateSnapBrickMountData(volinfo *volume.Volinfo, snapName string) (map[string]snapshot.BrickMountData, error) {
	nodeData := make(map[string]snapshot.BrickMountData)

	brickCount := 0
	for _, b := range volinfo.GetLocalBricks() {
		brickCount++
		mountRoot, err := volume.GetBrickMountRoot(b.Path)
		if err != nil {
			return nil, err
		}
		mountDir := b.Path[len(mountRoot):]
		mntInfo, err := volume.GetBrickMountInfo(mountRoot)
		if err != nil {
			log.WithError(err).WithField(
				"brick", b.Path,
			).Error("Failed to mount information")

			return nil, err
		}

		vG, err := lvm.GetVgName(mntInfo.FsName)
		if err != nil {

			log.WithError(err).WithField(
				"brick", b.Path,
			).Error("Failed to get vg name")

			return nil, err
		}
		devicePath := fmt.Sprintf("/dev/%s/%s_%s_%d", vG, "snap", snapName, brickCount)

		nodeID := strings.Replace(b.ID.String(), "-", "", -1)

		nodeData[b.String()] = snapshot.BrickMountData{
			MountDir:   mountDir,
			DevicePath: devicePath,
			FsType:     mntInfo.MntType,
			MntOpts:    updateMntOps(mntInfo.MntType, mntInfo.MntOpts),
			Path:       snapshotBrickCreate(snapName, volinfo.Name, nodeID, mountDir, brickCount),
		}
		// Store the results in transaction context. This will be consumed by
		// the node that initiated the transaction.

	}
	return nodeData, nil
}

func validateSnapCreate(c transaction.TxnCtx) error {
	var req api.SnapCreateReq
	var statusStr []string
	var err error
	var nodeData map[string]snapshot.BrickMountData
	var volinfo *volume.Volinfo
	if err = c.Get("req", &req); err != nil {
		return err
	}

	volinfo, err = volume.GetVolume(req.VolName)
	if err != nil {
		return err
	}
	if err = lvm.CommonPrevalidation(lvm.CreateCommand); err != nil {
		log.WithError(err).WithField(
			"command", lvm.CreateCommand,
		).Error("Failed to find lvm packages")
		return err
	}

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
	statusComptability := snapshot.CheckBricksCompatability(volinfo)
	if statusComptability != nil {
		log.WithError(err).WithField(
			"Bricks", statusStr,
		).Error("Bricks are not compatable")

		return errors.New("one or more brick is not compatable")
	}
	if nodeData, err = populateSnapBrickMountData(volinfo, req.SnapName); err != nil {

		return err
	}
	c.SetNodeResult(gdctx.MyUUID, snapshot.NodeDataTxnKey, &nodeData)
	//TODO Quorum check has to be implemented once we implement highly available snapshot
	return nil
}
func takeVolumeSnapshots(newVol, oldVol *volume.Volinfo) error {
	var wg sync.WaitGroup
	numBricks := len(oldVol.GetBricks())
	errCh := make(chan error, numBricks)
	for subvolCount, subvol := range oldVol.Subvols {
		for count, b := range subvol.Bricks {
			if !uuid.Equal(b.PeerID, gdctx.MyUUID) {
				continue
			}
			wg.Add(1)
			snapBrick := newVol.Subvols[subvolCount].Bricks[count]
			go brickSnapshot(errCh, &wg, snapBrick, b)

		}
	}
	err := error(nil)
	go func() {
		for i := range errCh {
			if i != nil && err == nil {
				//Return the first error from goroutines
				err = i
			}
		}
	}()

	wg.Wait()
	//Close will happen after executing all the go routines, so err should have populated by then
	//By the time return executes err will have the right value

	close(errCh)
	return err
}

func brickSnapshot(errCh chan error, wg *sync.WaitGroup, snapBrick, b brick.Brickinfo) {
	defer wg.Done()

	mountData := snapBrick.MountInfo
	length := len(b.Path) - len(mountData.Mountdir)
	mountRoot := b.Path[:length]
	mntInfo, err := volume.GetBrickMountInfo(mountRoot)
	if err != nil {
		errCh <- err
		return
	}

	log.WithFields(log.Fields{
		"mountDevice": mntInfo.FsName,
		"devicePath":  mountData.DevicePath,
		"Path":        b.Path,
	}).Debug("Running snapshot create command")

	if err := lvm.LVSnapshot(mntInfo.FsName, mountData.DevicePath); err != nil {
		log.WithFields(log.Fields{
			"mountDevice": mntInfo.FsName,
			"devicePath":  mountData.DevicePath,
			"Path":        b.Path,
		}).Error("Running snapshot create command failed")
		errCh <- err
		return
	}

	if err = lvm.UpdateFsLabel(mountData.DevicePath, mountData.FsType); err != nil {
		log.WithFields(log.Fields{
			"FsType": mountData.FsType,
			"Path":   b.Path,
		}).Error("Failed to update the label")
		errCh <- err
		return
	}
	errCh <- nil
	return
}

func takeSnapshots(c transaction.TxnCtx) error {
	var snapInfo snapshot.Snapinfo

	if err := c.Get("snapinfo", &snapInfo); err != nil {
		return err
	}

	snapVol := &snapInfo.SnapVolinfo
	volinfo, err := volume.GetVolume(snapInfo.ParentVolume)
	if err != nil {
		return err
	}
	err = takeVolumeSnapshots(snapVol, volinfo)
	return err

}

func createSnapSubvols(newVolinfo, origVolinfo *volume.Volinfo, nodeData map[string]snapshot.BrickMountData) error {
	var err error
	for idx, subvol := range origVolinfo.Subvols {
		subvolType := volume.SubvolTypeToString(subvol.Type)
		name := fmt.Sprintf("%s-%s-%d", newVolinfo.Name, strings.ToLower(subvolType), idx)
		s := volume.Subvol{
			Name: name,
			ID:   uuid.NewRandom(),
			Type: subvol.Type,
		}
		s.ArbiterCount = subvol.ArbiterCount
		s.DisperseCount = subvol.DisperseCount
		s.RedundancyCount = subvol.RedundancyCount
		s.ReplicaCount = subvol.ReplicaCount
		s.Subvols = subvol.Subvols
		//what is subvol ?
		{
			var bricks []api.BrickReq
			for _, brickinfo := range subvol.Bricks {
				mountData := nodeData[brickinfo.String()]
				peerID := brickinfo.PeerID.String()
				brick := api.BrickReq{
					PeerID: peerID,
					Type:   subvolType,
					Path:   mountData.Path,
				}

				bricks = append(bricks, brick)
			}
			s.Bricks, err = volume.NewBrickEntriesFunc(bricks, newVolinfo.Name, newVolinfo.ID)
			if err != nil {
				return err
			}
		}
		for count := 0; count < len(s.Bricks); count++ {
			key := subvol.Bricks[count].String()
			data := nodeData[key]
			s.Bricks[count].MountInfo = brick.MountInfo{
				Mountdir:   data.MountDir,
				DevicePath: data.DevicePath,
				FsType:     data.FsType,
				MntOpts:    data.MntOpts,
			}

		}
		newVolinfo.Subvols = append(newVolinfo.Subvols, s)

	}
	return nil
}

func createSnapinfo(c transaction.TxnCtx) error {
	var req api.SnapCreateReq
	ignoreOps := map[string]string{
		"features.quota":             "off",
		"features.inode-quota":       "off",
		"feature.deem-statfs":        "off",
		"features.quota-deem-statfs": "off",
		"bitrot-stub.bitrot":         "off",
		"replicate.self-heal-daemon": "off",
		"features.read-only":         "on",
		"features.uss":               "off",
	}

	nodeData := make(map[string]snapshot.BrickMountData)
	if err := c.Get("req", &req); err != nil {
		return err
	}

	volinfo, err := volume.GetVolume(req.VolName)
	if err != nil {
		return err
	}

	for _, node := range volinfo.Nodes() {
		tmp := make(map[string]snapshot.BrickMountData)
		err := c.GetNodeResult(node, snapshot.NodeDataTxnKey, &tmp)
		if err != nil {
			return err
		}
		for k, v := range tmp {
			nodeData[k] = v
		}
	}

	snapInfo := new(snapshot.Snapinfo)
	snapVolinfo := &snapInfo.SnapVolinfo
	duplicateVolinfo(volinfo, snapVolinfo)

	snapInfo.OptionChange = make(map[string]string)

	for key, value := range ignoreOps {
		snapInfo.OptionChange[key] = snapVolinfo.Options[key]
		snapVolinfo.Options[key] = value
	}

	snapVolinfo.State = volume.VolCreated
	snapVolinfo.GraphMap = volinfo.GraphMap
	snapVolinfo.ID = uuid.NewRandom()
	snapVolinfo.Name = req.SnapName
	snapVolinfo.VolfileID = "snaps/" + req.SnapName
	/*
		TODO
		For now disabling heal
	*/

	err = createSnapSubvols(snapVolinfo, volinfo, nodeData)
	if err != nil {
		log.WithFields(log.Fields{
			"snapshot":   snapVolinfo.Name,
			"volumeName": volinfo.Name,
		}).Error("Failed to create snap volinfo")

		return err
	}

	snapInfo.Description = req.Description
	snapInfo.ParentVolume = req.VolName
	/*
		Snapshot time would be a good addition ?
	*/

	err = c.Set("snapinfo", snapInfo)
	if err != nil {
		return err
	}
	return nil
}

func duplicateVolinfo(vol, v *volume.Volinfo) {

	v.Options = make(map[string]string)
	for key, value := range vol.Options {
		v.Options[key] = value
	}
	v.Transport = vol.Transport
	v.DistCount = vol.DistCount
	v.Type = vol.Type

	/*
		v.Checksum = 0
		v.Version = 0
	*/
	v.Auth = volume.VolAuth{
		Username: uuid.NewRandom().String(),
		Password: uuid.NewRandom().String(),
	}
	/*
	* Geo-replication cofig snapshot
	* Quota config snapshot
	* del barrier option
	 */
	return
}
func snapshotBrickCreate(snapName, volName, nodeID, mountDir string, brickCount int) string {
	snapDirPrefix := config.GetString("rundir") + "/snaps/"
	brickPath := fmt.Sprintf("%s%s/%s/%s/brick%d%s", snapDirPrefix, volName, nodeID, snapName, brickCount, mountDir)
	return brickPath
}

func validateOriginNodeSnapCreate(c transaction.TxnCtx) error {
	var req api.SnapCreateReq
	if err := c.Get("req", &req); err != nil {
		return err
	}
	if snapshot.ExistsFunc(req.SnapName) {
		return gderrors.ErrSnapExists
	}

	volinfo, err := volume.GetVolume(req.VolName)
	if err != nil {
		return err
	}

	if volinfo.State != volume.VolStarted {

		return errors.New("volume has not started")
	}
	barrierOp := volinfo.Options["features.barrier"]
	if err := c.Set("barrier-enabled", &barrierOp); err != nil {
		return err
	}
	if err := c.Set("originator-uuid", &gdctx.MyUUID); err != nil {
		return err
	}

	/*
		TODO
		*Geo-replication,
		*rebalance
		*tier daemon run check
		*check for hard-limit and soft-limit
		*auto-delete
	*/

	return nil
}

func registerSnapCreateStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"snap-create.OriginNodeValidate", validateOriginNodeSnapCreate},
		{"snap-create.Validate", validateSnapCreate},
		{"snap-create.CreateSnapinfo", createSnapinfo},
		{"snap-create.ActivateBarrier", activateBarrier},
		{"snap-create.TakeBrickSnapshots", takeSnapshots},
		{"snap-create.UndoBrickSnapshots", undoBrickSnapshots},
		{"snap-create.DeactivateBarrier", deactivateBarrier},
		{"snap-create.StoreSnapshot", storeSnapshot},
		{"snap-create.UndoStoreSnapshotOnCreate", undoStoreSnapshotOnCreate},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func snapshotCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := new(api.SnapCreateReq)
	logger := gdctx.GetReqLogger(ctx)
	var snapInfo snapshot.Snapinfo

	httpStatus, err := unmarshalSnapCreateRequest(req, r)
	if err != nil {
		logger.WithError(err).Error("Failed to unmarshal snaphot create request")
		restutils.SendHTTPError(ctx, w, httpStatus, err)
		return
	}
	if req.TimeStamp == true {
		t := time.Now().UTC()
		req.SnapName = req.SnapName + t.Format("_GMT_2006_01_02_15_04_05")
	}

	if !volume.IsValidName(req.SnapName) {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, gderrors.ErrInvalidSnapName)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, req.VolName, req.SnapName)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	vol, e := volume.GetVolume(req.VolName)
	if e != nil {
		status, err := restutils.ErrToStatusCode(e)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "snap-create.OriginNodeValidate",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "snap-create.Validate",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "snap-create.CreateSnapinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc:   "snap-create.ActivateBarrier",
			UndoFunc: "snap-create.DeactivateBarrier",
			Nodes:    txn.Nodes,
		},

		{
			DoFunc:   "snap-create.TakeBrickSnapshots",
			UndoFunc: "snap-create.UndoBrickSnapshots",
			Nodes:    txn.Nodes,
		},
		{
			DoFunc: "snap-create.DeactivateBarrier",
			Nodes:  txn.Nodes,
		},

		{
			DoFunc:   "snap-create.StoreSnapshot",
			UndoFunc: "snap-create.UndoStoreSnapshotOnCreate",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
	}
	err = txn.Ctx.Set("req", req)
	if err != nil {
		logger.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("snapshot create transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	txn.Ctx.Logger().WithField("SnapName", req.SnapName).Info("new snapshot created with status ", httpStatus)
	err = txn.Ctx.Get("snapinfo", &snapInfo)
	if err != nil {
		logger.WithError(err).Error("failed to get snap volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	resp := createSnapCreateResp(&snapInfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)

}

// createSnapCreateResp functions create resnse for rest utils
func createSnapCreateResp(snap *snapshot.Snapinfo) *api.SnapCreateResp {
	return (*api.SnapCreateResp)(createSnapInfoResp(snap))
}

func createSnapInfoResp(snap *snapshot.Snapinfo) *api.SnapInfo {
	var vinfo *api.VolumeInfo
	vinfo = volume.CreateVolumeInfoResp(&snap.SnapVolinfo)
	return &api.SnapInfo{
		VolInfo:       *vinfo,
		ParentVolName: snap.ParentVolume,
		Description:   snap.Description,
	}
}
