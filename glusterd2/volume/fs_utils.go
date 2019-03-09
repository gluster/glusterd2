package volume

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/gluster/glusterd2/glusterd2/brick"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	"golang.org/x/sys/unix"
)

const (
	volumeIDXattrKey = "trusted.glusterfs.volume-id"
)

//For now duplicating SizeInfo till we have a common package for both brick and volume

// SizeInfo represents sizing information.
type SizeInfo struct {
	Capacity uint64
	Used     uint64
	Free     uint64
}

func createSizeInfo(fstat *syscall.Statfs_t) *SizeInfo {
	var s SizeInfo
	if fstat != nil {
		s.Capacity = fstat.Blocks * uint64(fstat.Bsize)
		s.Free = fstat.Bfree * uint64(fstat.Bsize)
		s.Used = s.Capacity - s.Free
	}
	return &s
}

const fuseSuperMagic = 1702057286

//MountVolume mounts the gluster volume on a given mount point
func MountVolume(name string, mountpoint string, mntOptns string) error {
	// NOTE: Why do it this way ?
	// * Libgfapi leaks memory on unmount.
	// * Glusterfs volumes cannot be mounted using syscall.Mount()

	shost, sport, err := net.SplitHostPort(config.GetString("clientaddress"))
	if err != nil {
		return err
	}

	if shost == "" {
		shost = "127.0.0.1"
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" --volfile-server %s", shost))
	buffer.WriteString(fmt.Sprintf(" --volfile-server-port %s", sport))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", name))

	mountpointWithoutSlash := strings.Trim(strings.Replace(mountpoint, "/", "-", -1), "-")
	logfilepath := path.Join(config.GetString("logdir"), "glusterfs", mountpointWithoutSlash)
	buffer.WriteString(" --log-file " + logfilepath)

	buffer.WriteString(mntOptns)
	buffer.WriteString(" " + mountpoint)

	args := strings.Fields(buffer.String())
	cmd := exec.Command("glusterfs", args...)
	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait() // glusterfs daemonizes itself
}

//UsageInfo gives the size information of a gluster volume
func UsageInfo(volname string) (*SizeInfo, error) {

	tempDir, err := ioutil.TempDir(config.GetString("rundir"), "gd2mount")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempDir)

	if err := MountVolume(volname, tempDir, " --read-only "); err != nil {
		return nil, err
	}
	defer syscall.Unmount(tempDir, syscall.MNT_FORCE)

	var fstat syscall.Statfs_t
	if err := syscall.Statfs(tempDir, &fstat); err != nil {
		return nil, err
	}

	if fstat.Type != fuseSuperMagic {
		// Do a crude check if mountpoint is a glusterfs mount
		return nil, errors.New("not FUSE mount")
	}

	return createSizeInfo(&fstat), nil
}

//Mntent is used to reprsent a state of a mount entry in mtab
type Mntent struct {
	FsName  string
	MntDir  string
	MntType string
	MntOpts string
	// excluded mnt_freq and mnt_passno
}

// See `man getmntent`
var mtabReplacer = strings.NewReplacer("\\040", " ", "\\011", "\t", "\\012", "\n", "\\134", "\\")

func readMountEntry(entry string) *Mntent {
	f := strings.Fields(entry)
	if len(f) != 6 {
		return nil
	}

	for i := 0; i < 4; i++ {
		f[i] = mtabReplacer.Replace(f[i])
	}

	return &Mntent{
		FsName:  f[0],
		MntDir:  f[1],
		MntType: f[2],
		MntOpts: f[3],
	}
}

//GetMounts returns all the mount point entries from /proc/mounts
func GetMounts() ([]*Mntent, error) {

	content, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return nil, err
	}

	var l []*Mntent

	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		m := readMountEntry(scanner.Text())
		if m != nil {
			l = append(l, m)
		}
	}

	return l, nil
}

//IsMountExist return success when mount point already exist
//If mtab values is given, it does high level validation of mount
//Else it will skip device verification and just does xattr validation
func IsMountExist(b *brick.Brickinfo, iD uuid.UUID, mtab []*Mntent) bool {

	if _, err := os.Lstat(b.Path); err != nil {
		return false
	}

	data := make([]byte, 16)
	sz, err := syscall.Getxattr(b.Path, volumeIDXattrKey, data)
	if err != nil || sz <= 0 {
		return false
	}

	//Check for little or big endian ?
	if !uuid.Equal(iD, data[:sz]) {
		return false
	}

	if mtab == nil {
		//Skip high level mount validation
		return true
	}

	mountData := b.MountInfo
	mountRoot := strings.TrimSuffix(b.Path, mountData.BrickDirSuffix)
	for _, entry := range mtab {
		if entry.MntDir == mountRoot {
			if entry.MntType != mountData.FsType {
				return false
			}
			devicePath, err := os.Readlink(mountData.DevicePath)
			if err != nil {
				return false
			}
			deviceMapper, err := os.Readlink(entry.FsName)
			if err != nil {
				return false
			}
			if deviceMapper == devicePath {
				return true
			}
			//No need to continue as mount root already processed
			break
		}
	}
	return false
}

//UmountBrickDirectory does an umount of the path
func UmountBrickDirectory(path string) error {
	return syscall.Unmount(path, syscall.MNT_FORCE)
}

//MountBrickDirectory creates the directory strcture for bricks
func MountBrickDirectory(vol *Volinfo, brickinfo *brick.Brickinfo, mtab []*Mntent) error {

	provisionType := brickinfo.PType
	if !(provisionType.IsAutoProvisioned() || provisionType.IsSnapshotProvisioned()) {
		return nil
	}

	mountData := brickinfo.MountInfo
	mountRoot := strings.TrimSuffix(brickinfo.Path, mountData.BrickDirSuffix)
	//Because of abnormal shutdown of the brick, mount point might already be existing
	if IsMountExist(brickinfo, vol.ID, mtab) {
		log.WithFields(log.Fields{
			"mountRoot":  mountRoot,
			"fsType":     mountData.FsType,
			"devicePath": mountData.DevicePath,
		}).Debug("Mount point already exist")
		return nil
	}

	if err := os.MkdirAll(mountRoot, os.ModeDir|os.ModePerm); err != nil {
		log.WithError(err).Error("Failed to create snapshot directory ", brickinfo.String())
		return err
	}

	if err := MountDirectory(mountRoot, brickinfo.MountInfo); err != nil {
		log.WithError(err).WithFields(log.Fields{"brickPath": brickinfo.String(),
			"mountRoot": mountRoot}).Error("Failed to mount snapshot directory")

		return err
	}

	if err := unix.Setxattr(brickinfo.Path, volumeIDXattrKey, vol.ID, 0); err != nil {
		log.WithError(err).WithFields(log.Fields{"brickPath": brickinfo.Path,
			"xattr": volumeIDXattrKey}).Error("setxattr failed")
		return err
	}

	return nil
}

//MountDirectory will mount the bricks to the given path
func MountDirectory(mountPath string, mountData brick.MountInfo) error {
	// Use syscall.Mount command to mount the bricks
	return utils.ExecuteCommandRun("mount", "-o", mountData.MntOpts, mountData.DevicePath, mountPath)
}

//StopBrick terminate the process and umount the brick directory
func StopBrick(b brick.Brickinfo, logger log.FieldLogger) error {
	if err := b.TerminateBrick(); err != nil {
		if err = b.StopBrick(logger); err != nil {
			return err
		}
	}
	return UmountBrick(b)
}

//UmountBrick will umount the brick directory
func UmountBrick(b brick.Brickinfo) error {
	//TODO Validate mount point before umount
	var err error

	length := len(b.Path) - len(b.MountInfo.BrickDirSuffix)
	for j := 0; j < 3; j++ {
		err = UmountBrickDirectory(b.Path[:length])
		if err == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	return err
}

//MountVolumeBricks will mount local bricks of a volume
func MountVolumeBricks(volinfo *Volinfo, skipError bool) error {
	mtab, err := GetMounts()
	if err != nil {
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {
		err := MountBrickDirectory(volinfo, &b, mtab)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"brickPath": b.Path}).Error(gderrors.ErrBrickMountFailed)
			if !skipError {
				return err
			}
			continue
		}
	}
	return nil
}
