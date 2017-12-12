package volumecommands

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/pkg/api"

	config "github.com/spf13/viper"
)

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

func volumeUsage(volname string) (*api.SizeInfo, error) {

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
		return nil, errors.New("Not FUSE mount")
	}

	var v api.SizeInfo
	v.Capacity = fstat.Blocks * uint64(fstat.Bsize)
	v.Free = fstat.Bfree * uint64(fstat.Bsize)
	v.Used = v.Capacity - v.Free

	return &v, nil
}
