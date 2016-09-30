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
	"github.com/pborman/uuid"

	config "github.com/spf13/viper"
)

const (
	// TODO: Remove hardcoding
	glusterfsd = "/usr/sbin/glusterfsd"
)

// Brickinfo represents the information of a brick
// TODO: Move this into Brick struct ?
type Brickinfo struct {
	Hostname string
	Path     string
	ID       uuid.UUID
}

type Brick struct {
	// Externally consumable using methods of Daemon interface
	binarypath     string
	args           string
	socketfilepath string
	pidfilepath    string

	// For internal use
	brickinfo Brickinfo
	volName   string // Introduce this in Brickinfo itself ?
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
	//TODO: For now, getting next available port. Use portmap ?
	brickPort := strconv.Itoa(GetNextAvailableFreePort())
	//TODO: Passing volfile directly for now.
	//Change this once we have volfile fetch support in GD2.
	volfile := utils.GetBrickVolFilePath(b.volName, b.brickinfo.Hostname, b.brickinfo.Path)

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" -f %s", volfile))
	buffer.WriteString(fmt.Sprintf(" -p %s", b.PidFile()))
	buffer.WriteString(fmt.Sprintf(" -S %s", b.SocketFile()))
	buffer.WriteString(fmt.Sprintf(" --brick-name %s", b.brickinfo.Path))
	buffer.WriteString(fmt.Sprintf(" --brick-port %s", brickPort))
	buffer.WriteString(fmt.Sprintf(" -l %s", logFile))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *-posix.glusterd-uuid=%s", gdctx.MyUUID))
	buffer.WriteString(fmt.Sprintf(" --xlator-option %s-server.listen-port=%s", b.volName, brickPort))

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
	volumedir := utils.GetVolumeDir(b.volName)
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
func NewDaemon(volName string, binfo Brickinfo) (*Brick, error) {
	brickObject := &Brick{binarypath: glusterfsd, brickinfo: binfo, volName: volName}
	return brickObject, nil
}
