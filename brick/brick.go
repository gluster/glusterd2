package brick

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	config "github.com/spf13/viper"
)

const (
	// TODO: Remove hardcoding
	glusterfsd = "/usr/sbin/glusterfsd"
)

type Brick struct {
	// Externally consumable using methods of Daemon interface
	binarypath     string
	args           string
	socketfilepath string
	pidfilepath    string

	// For internal use
	brickinfo volume.Brickinfo
	volinfo   *volume.Volinfo
	port      int
}

func (b *Brick) Name() string {
	return "glusterfsd"
}

func (b *Brick) Path() string {
	return b.binarypath
}

func (b *Brick) Args() string {
	if b.args != "" {
		return b.args
	}

	brickPathWithoutSlashes := strings.Trim(strings.Replace(b.brickinfo.Path, "/", "-", -1), "-")
	logFile := path.Join(config.GetString("logdir"), "glusterfs", "bricks", fmt.Sprintf("%s.log", brickPathWithoutSlashes))
	volFileId := fmt.Sprintf("%s.%s.%s", b.volinfo.Name, b.brickinfo.Hostname, brickPathWithoutSlashes)
	//TODO: For now, getting next available port. Use portmap ?
	brickPort := strconv.Itoa(GetNextAvailableFreePort())

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" -s %s", b.brickinfo.Hostname))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", volFileId))
	buffer.WriteString(fmt.Sprintf(" -p %s", b.PidFile()))
	buffer.WriteString(fmt.Sprintf(" -S %s", b.SocketFile()))
	buffer.WriteString(fmt.Sprintf(" --brick-name %s", b.brickinfo.Path))
	buffer.WriteString(fmt.Sprintf(" --brick-port %s", brickPort))
	buffer.WriteString(fmt.Sprintf(" -l %s", logFile))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *-posix.glusterd-uuid=%s", gdctx.MyUUID))
	buffer.WriteString(fmt.Sprintf(" --xlator-option %s-server.listen-port=%s", b.volinfo.Name, brickPort))

	b.args = buffer.String()
	return b.args
}

func (b *Brick) SocketFile() string {

	if b.socketfilepath != "" {
		return b.socketfilepath
	}

	// This looks a little convoluted but just doing what gd1 did...

	// First we form a fake path to the socket file
	// Example: /var/lib/glusterd/vols/<vol-name>/run/<host-name>-<brick-path>
	brickPathWithoutSlashes := strings.Trim(strings.Replace(b.brickinfo.Path, "/", "-", -1), "-")
	fakeSockFileName := fmt.Sprintf("%s-%s", b.brickinfo.Hostname, brickPathWithoutSlashes)
	volumedir := utils.GetVolumeDir(b.volinfo.Name)
	fakeSockFilePath := path.Join(volumedir, "run", fakeSockFileName)

	// Then md5sum of the above path shall be the name of socket file.
	// Example: /var/run/gluster/<md5sum-hash>.socket
	checksum_data := []byte(fakeSockFilePath)
	glusterdSockDir := path.Join(config.GetString("rundir"), "gluster")
	b.socketfilepath = fmt.Sprintf("%s/%x.socket", glusterdSockDir, md5.Sum(checksum_data))

	return b.socketfilepath
}

func (b *Brick) PidFile() string {

	if b.pidfilepath != "" {
		return b.pidfilepath
	}

	rundir := config.GetString("rundir")
	brickPathWithoutSlashes := strings.Trim(strings.Replace(b.brickinfo.Path, "/", "-", -1), "-")
	pidfilename := fmt.Sprintf("%s-%s.pid", b.brickinfo.Hostname, brickPathWithoutSlashes)
	b.pidfilepath = path.Join(rundir, "gluster", pidfilename)

	return b.pidfilepath
}

// Returns a new instance of Brick type which implements the Daemon interface
func NewDaemon(vinfo *volume.Volinfo, binfo volume.Brickinfo) (*Brick, error) {
	brickObject := &Brick{binarypath: glusterfsd, brickinfo: binfo, volinfo: vinfo}
	return brickObject, nil
}
