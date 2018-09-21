package snapshot

import (
	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/snapshot/lvm"
	"github.com/gluster/glusterd2/glusterd2/volume"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	volumeIDXattrKey = "trusted.glusterfs.volume-id"
	gfidXattrKey     = "trusted.gfid"
)

//BrickMountData contains information about mount point
type BrickMountData struct {
	//BrickDirSuffix is the directory path after brick mount root
	BrickDirSuffix string
	//DevicePath of brick mount
	DevicePath string
	//FsType is the file system type
	FsType string
	//MntOpts is mount option
	MntOpts string
	//Path need to be calculated from each node as there could be multiple glusterd's running on the same node
	Path string
}

var (
	//NodeDataTxnKey is used for storing the status
	NodeDataTxnKey = "brickmountstatus"
	//NodeConfigTxnKey is used for storing the config status
	NodeConfigTxnKey = "snapshotconfigstatus"
)

func getOnlineOfflineBricks(vol *volume.Volinfo, online bool) ([]brick.Brickinfo, error) {
	var brickinfos []brick.Brickinfo

	if vol.State == volume.VolStopped {
		//If volume is not started, We will assume all bricks are stopped.
		if online == true {
			return brickinfos, nil
		}
		return vol.GetLocalBricks(), nil
	}
	brickStatuses, err := volume.CheckBricksStatus(vol)
	if err != nil {
		return brickinfos, err
	}

	for _, brick := range brickStatuses {
		if brick.Online == online {
			brickinfos = append(brickinfos, brick.Info)
		}
	}
	return brickinfos, nil
}

//GetOfflineBricks will return slice of brickinfos that are offline on the local node
func GetOfflineBricks(vol *volume.Volinfo) ([]brick.Brickinfo, error) {
	return getOnlineOfflineBricks(vol, false)
}

//GetOnlineBricks will return slice of brickinfos that are online on the local node
func GetOnlineBricks(vol *volume.Volinfo) ([]brick.Brickinfo, error) {
	return getOnlineOfflineBricks(vol, true)
}

//ActivateDeactivateFunc uses to activate and deactivate
func ActivateDeactivateFunc(snapinfo *Snapinfo, b []brick.Brickinfo, activate bool, logger log.FieldLogger) error {
	volinfo := &snapinfo.SnapVolinfo
	switch volinfo.State == volume.VolStarted {
	case true:
		if len(b) == 0 {
			return nil
		}
	}
	mtab, err := volume.GetMounts()
	if err != nil {
		return err
	}
	for i := 0; i < len(b); i++ {

		if !uuid.Equal(b[i].PeerID, gdctx.MyUUID) {
			continue
		}

		if activate == true {
			if err := volume.MountBrickDirectory(volinfo, &b[i], mtab); err != nil {
				return err
			}
			if err := b[i].StartBrick(logger); err != nil {
				if err == gderrors.ErrProcessAlreadyRunning {
					continue
				}
				return err
			}

		} else {
			if err := volume.StopBrick(b[i], logger); err != nil {
				return err
			}
		}

	}

	return nil
}

//CheckBricksFsCompatability will verify the brickes are compatable
func CheckBricksFsCompatability(volinfo *volume.Volinfo) []string {

	var paths []string
	for _, brick := range volinfo.GetLocalBricks() {
		if lvm.FsCompatibleCheck(brick.Path) != true {
			paths = append(paths, brick.String())
		}
	}
	return paths

}

//CheckBricksSizeCompatability will verify the device has enough space
func CheckBricksSizeCompatability(volinfo *volume.Volinfo) []string {

	var paths []string
	for _, brick := range volinfo.GetLocalBricks() {
		if lvm.SizeCompatibleCheck(brick.Path) != true {
			paths = append(paths, brick.String())
		}
	}
	return paths

}
