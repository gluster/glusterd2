package volumecommands

import (
	"bytes"
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
	cmd.Wait() // glusterfs daemonizes itself

	return nil
}

func volumeUsage(volname string) (*api.SizeInfo, error) {

	var v api.SizeInfo
	tempDir, err := ioutil.TempDir(config.GetString("rundir"), "gd2mount")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempDir)

	if err := mountVolume("test", tempDir); err != nil {
		return nil, err
	}
	defer syscall.Unmount(tempDir, syscall.MNT_FORCE)

	var fstat syscall.Statfs_t
	if err := syscall.Statfs(tempDir, &fstat); err != nil {
		return nil, err
	}

	v.Capacity = fstat.Blocks * uint64(fstat.Bsize)
	v.Free = fstat.Bfree * uint64(fstat.Bsize)
	v.Used = v.Capacity - v.Free

	return &v, nil
}
