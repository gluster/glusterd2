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
	"strings"
	"syscall"

	config "github.com/spf13/viper"
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

func mountVolume(name string, mountpoint string) error {
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
	buffer.WriteString(" --log-file /dev/null")
	buffer.WriteString(" --read-only ")
	buffer.WriteString(mountpoint)

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

	if err := mountVolume(volname, tempDir); err != nil {
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
