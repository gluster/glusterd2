package brick

import (
	"crypto/md5"
	"fmt"
	"path"
	"strings"

	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"
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

	// Just doing whatever glusterd1 does
	volumedir := utils.GetVolumeDir(b.volinfo.Name)
	brickPathWithoutSlashes := strings.Replace(b.brickinfo.Path, "/", "-", -1)
	fakeSockFileName := fmt.Sprintf("%s-%s", b.brickinfo.Hostname, brickPathWithoutSlashes)
	fakeSockFilePath := path.Join(volumedir, "run", fakeSockFileName)
	checksum_data := []byte(fakeSockFilePath)
	b.socketfilepath = fmt.Sprintf("%s/%x.socket", utils.GlusterdSockDir, md5.Sum(checksum_data))

	return b.socketfilepath
}

func (b *Brick) PidFile() string {

	if b.pidfilepath != "" {
		return b.pidfilepath
	}

	// TODO: A more "standard location" to place pidfiles in is /var/run/
	// For now, following whatever glusterd1 does
	volumedir := utils.GetVolumeDir(b.volinfo.Name)
	brickPathWithoutSlashes := strings.Replace(b.brickinfo.Path, "/", "-", -1)
	pidfilename := fmt.Sprintf("%s-%s.pid", b.brickinfo.Hostname, brickPathWithoutSlashes)
	b.pidfilepath = path.Join(volumedir, "run", pidfilename)

	// It is assumed that the responsibility of this function is to only
	// generate and return the pid file path and not to create it ?
	return b.pidfilepath
}

// Returns a new instance of Brick type which implements the Daemon interface
func NewDaemon(vinfo *volume.Volinfo, binfo volume.Brickinfo) (*Brick, error) {
	brickObject := &Brick{binarypath: glusterfsd, brickinfo: binfo, volinfo: vinfo}
	return brickObject, nil
}
