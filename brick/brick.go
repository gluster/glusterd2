package brick

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/utils"
	"github.com/pborman/uuid"

	config "github.com/spf13/viper"
)

const (
	glusterfsdBin = "glusterfsd"
)

// Brickinfo represents the information of a brick
// TODO: Move this into Brick struct ?
type Brickinfo struct {
	Hostname string
	Path     string
	ID       uuid.UUID
}

// Brickstatus represents real-time status of the brick
// TODO: Consolidate fields of Brickinfo, Brickstatus and Brick structs
type Brickstatus struct {
	Hostname string
	Path     string
	ID       uuid.UUID
	Online   bool
	Pid      int
	// TODO: Add other fields like filesystem type, statvfs output etc.
}

// Brick type represents information about the brick daemon
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

// Name returns human-friendly name of the brick process. This is used for logging.
func (b *Brick) Name() string {
	return "glusterfsd"
}

// Path returns absolute path to the binary of brick process
func (b *Brick) Path() string {
	return b.binarypath
}

// Args returns arguments to be passed to brick process during spawn.
func (b *Brick) Args() string {

	brickPathWithoutSlashes := strings.Trim(strings.Replace(b.brickinfo.Path, "/", "-", -1), "-")
	logFile := path.Join(config.GetString("logdir"), "glusterfs", "bricks", fmt.Sprintf("%s.log", brickPathWithoutSlashes))
	//TODO: For now, getting next available port. Use portmap ?
	brickPort := strconv.Itoa(GetNextAvailableFreePort())
	//TODO: Passing volfile directly for now.
	//Change this once we have volfile fetch support in GD2.
	volfile := utils.GetBrickVolFilePath(b.volName, b.brickinfo.Hostname, b.brickinfo.Path)

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" --volfile %s", volfile))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", b.volName))
	buffer.WriteString(fmt.Sprintf(" -p %s", b.PidFile()))
	buffer.WriteString(fmt.Sprintf(" -S %s", b.SocketFile()))
	buffer.WriteString(fmt.Sprintf(" --brick-name %s", b.brickinfo.Path))
	buffer.WriteString(fmt.Sprintf(" --brick-port %s", brickPort))
	buffer.WriteString(fmt.Sprintf(" -l %s", logFile))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *-posix.glusterd-uuid=%s", gdctx.MyUUID))
	buffer.WriteString(fmt.Sprintf(" --xlator-option %s-server.transport.socket.listen-port=%s", b.volName, brickPort))

	b.args = buffer.String()
	return b.args
}

// SocketFile returns path to the brick socket file used for IPC.
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
	checksumData := []byte(fakeSockFilePath)
	glusterdSockDir := path.Join(config.GetString("rundir"), "gluster")
	b.socketfilepath = fmt.Sprintf("%s/%x.socket", glusterdSockDir, md5.Sum(checksumData))

	return b.socketfilepath
}

// PidFile returns path to the pid file of the brick process
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

// NewDaemon returns a new instance of Brick type which implements the Daemon interface
func NewDaemon(volName string, binfo Brickinfo) (*Brick, error) {
	path, e := exec.LookPath(glusterfsdBin)
	if e != nil {
		return nil, e
	}
	brickObject := &Brick{binarypath: path, brickinfo: binfo, volName: volName}
	return brickObject, nil
}
