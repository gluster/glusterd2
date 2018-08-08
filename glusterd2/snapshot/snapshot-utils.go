package snapshot

import (
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/snapshot/lvm"
	"github.com/gluster/glusterd2/glusterd2/volume"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const (
	volumeIDXattrKey = "trusted.glusterfs.volume-id"
	gfidXattrKey     = "trusted.gfid"
)

//BrickMountData contains information about mount point
type BrickMountData struct {
	//MountDir is the directory path after brick mount
	MountDir string
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
)

//UmountSnapBrickDirectory does an umount of the path
func UmountSnapBrickDirectory(path string) error {
	err := syscall.Unmount(path, syscall.MNT_FORCE)
	return err
}

//IsMountExist return success when mount point already exist
func IsMountExist(brickPath string, iD uuid.UUID) bool {

	if _, err := os.Lstat(brickPath); err != nil {
		return false
	}

	data := make([]byte, 16)
	sz, err := syscall.Getxattr(brickPath, volumeIDXattrKey, data)
	if err != nil || sz <= 0 {
		return false
	}

	//Check for little or big endian ?
	if uuid.Equal(iD, data[:sz]) {
		return true
	}
	//TODO add mor checks to confim the mount point, like device verification
	return false
}

//MountSnapBrickDirectory creates the directory strcture for snap bricks
func MountSnapBrickDirectory(vol *volume.Volinfo, brickinfo *brick.Brickinfo) error {

	mountData := brickinfo.MountInfo
	mountRoot := strings.TrimSuffix(brickinfo.Path, mountData.Mountdir)
	//Because of abnormal shutdown of the brick, mount point might already be existing
	if IsMountExist(brickinfo.Path, vol.ID) {
		return nil
	}

	if err := os.MkdirAll(mountRoot, os.ModeDir|os.ModePerm); err != nil {
		log.WithError(err).Error("Failed to create snapshot directory ", brickinfo.String())
		return err
	}
	/*
	   TODO
	   *Move to snapshot package as it has no lvm related coomands.
	*/

	if err := lvm.MountSnapshotDirectory(mountRoot, brickinfo.MountInfo); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"brickPath": brickinfo.String(),
			"mountRoot": mountRoot}).Error("Failed to mount snapshot directory")

		return err
	}

	err := unix.Setxattr(brickinfo.Path, volumeIDXattrKey, vol.ID, 0)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"brickPath": brickinfo.Path,
			"xattr":     volumeIDXattrKey}).Error("setxattr failed")
		return err
	}

	return nil
}

func getOnlineOfflineBricks(vol *volume.Volinfo, online bool) ([]brick.Brickinfo, error) {
	var brickinfos []brick.Brickinfo

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
	for i := 0; i < len(b); i++ {

		if !uuid.Equal(b[i].PeerID, gdctx.MyUUID) {
			continue
		}

		if activate == true {
			if err := MountSnapBrickDirectory(volinfo, &b[i]); err != nil {
				return err
			}
			if err := b[i].StartBrick(logger); err != nil {
				if err == gderrors.ErrProcessAlreadyRunning {
					continue
				}
				return err
			}

		} else {
			if err := StopBrick(b[i], logger); err != nil {
				return err
			}
		}

	}

	return nil
}

//CheckBricksCompatability will verify the brickes are lvm compatable
func CheckBricksCompatability(volinfo *volume.Volinfo) []string {

	var paths []string
	for _, subvol := range volinfo.Subvols {
		for _, brick := range subvol.Bricks {
			if lvm.IsThinLV(brick.Path) != true {
				paths = append(paths, brick.String())
			}
		}
	}
	return paths

}

//StopBrick terminate the process and umount the brick directory
func StopBrick(b brick.Brickinfo, logger log.FieldLogger) error {
	var err error

	if err = b.TerminateBrick(); err != nil {
		if err = b.StopBrick(logger); err != nil {
			return err
		}
	}

	length := len(b.Path) - len(b.MountInfo.Mountdir)
	for j := 0; j < 3; j++ {

		err = UmountSnapBrickDirectory(b.Path[:length])
		if err == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	return err

}

//UmountBrick will umount the brick directory
func UmountBrick(b brick.Brickinfo) error {
	var err error

	length := len(b.Path) - len(b.MountInfo.Mountdir)
	for j := 0; j < 3; j++ {

		err = UmountSnapBrickDirectory(b.Path[:length])
		if err == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	return err

}
