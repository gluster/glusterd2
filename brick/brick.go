package brick

import (
	"crypto/md5"
	"fmt"
	"path"
	"strings"

	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	config "github.com/spf13/viper"
)

const (
	// TODO: Remove hardcoding
	//	glusterfsd = "/usr/local/sbin/glusterfsd"
	glusterfsd = "/usr/bin/sleep"
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

	return "60"
	//return b.args
}

func (b *Brick) SocketFile() string {

	if b.socketfilepath != "" {
		return b.socketfilepath
	}

	// This looks a little convoluted but just doing what gd1 did...

	// First we form a fake path to the socket file
	// Example: /var/lib/glusterd/vols/<vol-name>/run/<host-name>-<brick-path>
	brickPathWithoutSlashes := strings.Replace(b.brickinfo.Path, "/", "-", -1)
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
	brickPathWithoutSlashes := strings.Replace(b.brickinfo.Path, "/", "-", -1)
	pidfilename := fmt.Sprintf("%s%s.pid", b.brickinfo.Hostname, brickPathWithoutSlashes)
	b.pidfilepath = path.Join(rundir, pidfilename)

	return b.pidfilepath
}

// Returns a new instance of Brick type which implements the Daemon interface
func NewDaemon(vinfo *volume.Volinfo, binfo volume.Brickinfo) (*Brick, error) {
	brickObject := &Brick{binarypath: glusterfsd, brickinfo: binfo, volinfo: vinfo}
	return brickObject, nil
}
