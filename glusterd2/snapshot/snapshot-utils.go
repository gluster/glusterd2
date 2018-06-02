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
}

var (
	//NodeDataTxnKey is used for storing the status
	NodeDataTxnKey = "brickmountstatus"
)

//UmountSnapBrickDirectory does an umount of the path
func UmountSnapBrickDirectory(path string) error {
	//	_, err := exec.Command("umount", "-f", path).Output()
	err := syscall.Unmount(path, syscall.MNT_FORCE)
	return err
}

//MountSnapBrickDirectory creates the directory strcture for snap bricks
func MountSnapBrickDirectory(vol *volume.Volinfo, brickinfo *brick.Brickinfo) error {

	mountData := brickinfo.MountInfo
	mountRoot := strings.TrimSuffix(brickinfo.Path, mountData.Mountdir)
	if err := os.MkdirAll(mountRoot, os.ModeDir|os.ModePerm); err != nil {
		log.WithError(err).Error("Failed to create snapshot directory ", brickinfo.String())
		return err
	}
	/*
	   TODO
	   *Move to snapshot package as it has no lvm related coomands.
	   *Handle already mounted path, eg: using start when a brick is down, mostly path could be mounted.
	*/

	if err := lvm.MountSnapshotDirectory(mountRoot, brickinfo.MountInfo); err != nil {
		log.WithError(err).Error("Failed to mount snapshot directory ", brickinfo.String())
		return err
	}

	err := unix.Setxattr(brickinfo.Path, volumeIDXattrKey, vol.ID, 0)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error(),
			"brickPath": brickinfo.Path,
			"xattr":     volumeIDXattrKey}).Error("setxattr failed")
		return err
	}

	return nil
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
