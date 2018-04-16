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

const (
	//NodeDataTxnKey is used for storing the status
	NodeDataTxnKey string = "brickmountstatus"
	//SnapDirPrefix contains the prefix of snapshot brick
	SnapDirPrefix string = "/var/run/gluster/snaps/"
)

//UmountSnapBrickDirectory does an umount of the path
func UmountSnapBrickDirectory(path string) error {
	//	_, err := exec.Command("umount", "-f", path).Output()
	err := syscall.Unmount(path, syscall.MNT_FORCE)
	return err
}

//MountSnapBrickDirectory creates the directory strcture for snap bricks
func MountSnapBrickDirectory(vol *volume.Volinfo, brickinfo *brick.Brickinfo) error {

	mountRoot := strings.TrimSuffix(brickinfo.Path, brickinfo.Mountdir)
	if err := os.MkdirAll(mountRoot, os.ModeDir|os.ModePerm); err != nil {
		log.WithError(err).Error("Failed to create snapshot directory ", brickinfo.String())
		return err
	}
	/*
	   TODO
	   *Move to snapshot package as it has no lvm related coomands
	   *Handle already mounted path, eg: using start when a brick is down, mostly path could be mounted
	*/

	if err := lvm.MountSnapshotDirectory(mountRoot, brickinfo); err != nil {
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
func ActivateDeactivateFunc(snapinfo *Snapinfo, b []brick.Brickinfo, activate bool) error {
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
			if err := b[i].StartBrick(); err != nil {
				return err
			}

		} else {
			var err error
			if err = b[i].StopBrick(); err != nil {
				return err
			}

			length := len(b[i].Path) - len(b[i].Mountdir)
			for j := 0; j < 3; j++ {

				err = UmountSnapBrickDirectory(b[i].Path[:length])
				if err == nil {
					break
				}
				time.Sleep(3 * time.Second)
			}
			if err != nil {
				return err
			}
		}

	}

	return nil
}
