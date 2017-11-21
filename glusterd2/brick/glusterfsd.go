package brick

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/cespare/xxhash"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/pmap"
	"github.com/gluster/glusterd2/pkg/utils"

	config "github.com/spf13/viper"
)

const (
	glusterfsdBin = "glusterfsd"
)

// Glusterfsd type represents information about the brick daemon
type Glusterfsd struct {
	// Externally consumable using methods of Glusterfsd interface
	binarypath     string
	args           string
	socketfilepath string
	pidfilepath    string

	// For internal use
	brickinfo Brickinfo
}

// Name returns human-friendly name of the brick process. This is used for logging.
func (b *Glusterfsd) Name() string {
	return "glusterfsd"
}

// Path returns absolute path to the binary of brick process
func (b *Glusterfsd) Path() string {
	return b.binarypath
}

// Args returns arguments to be passed to brick process during spawn.
func (b *Glusterfsd) Args() string {

	brickPathWithoutSlashes := strings.Trim(strings.Replace(b.brickinfo.Path, "/", "-", -1), "-")

	logFile := path.Join(config.GetString("logdir"), "glusterfs", "bricks", fmt.Sprintf("%s.log", brickPathWithoutSlashes))

	brickPort := strconv.Itoa(pmap.AssignPort(0, b.brickinfo.Path))

	volFileID := b.brickinfo.VolumeName + "." + gdctx.MyUUID.String() + "." + brickPathWithoutSlashes

	shost, sport, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if shost == "" {
		shost = "127.0.0.1"
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" --volfile-server %s", shost))
	buffer.WriteString(fmt.Sprintf(" --volfile-server-port %s", sport))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", volFileID))
	buffer.WriteString(fmt.Sprintf(" -p %s", b.PidFile()))
	buffer.WriteString(fmt.Sprintf(" -S %s", b.SocketFile()))
	buffer.WriteString(fmt.Sprintf(" --brick-name %s", b.brickinfo.Path))
	buffer.WriteString(fmt.Sprintf(" --brick-port %s", brickPort))
	buffer.WriteString(fmt.Sprintf(" -l %s", logFile))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *-posix.glusterd-uuid=%s", gdctx.MyUUID))
	buffer.WriteString(fmt.Sprintf(" --xlator-option %s-server.transport.socket.listen-port=%s", b.brickinfo.VolumeName, brickPort))

	b.args = buffer.String()
	return b.args
}

// SocketFile returns path to the brick socket file used for IPC.
func (b *Glusterfsd) SocketFile() string {

	if b.socketfilepath != "" {
		return b.socketfilepath
	}

	// This looks a little convoluted but just doing what gd1 did...

	// First we form a fake path to the socket file
	// Example: /var/lib/glusterd/vols/<vol-name>/run/<host-name>-<brick-path>
	brickPathWithoutSlashes := strings.Trim(strings.Replace(b.brickinfo.Path, "/", "-", -1), "-")
	// FIXME: The brick can no longer clean this up on clean shut down
	fakeSockFileName := fmt.Sprintf("%s-%s", b.brickinfo.NodeID.String(), brickPathWithoutSlashes)
	volumedir := utils.GetVolumeDir(b.brickinfo.VolumeName)
	fakeSockFilePath := path.Join(volumedir, "run", fakeSockFileName)

	// Then xxhash of the above path shall be the name of socket file.
	// Example: /var/run/gluster/<xxhash-hash>.socket
	glusterdSockDir := path.Join(config.GetString("rundir"), "gluster")
	b.socketfilepath = fmt.Sprintf("%s/%x.socket", glusterdSockDir, xxhash.Sum64String(fakeSockFilePath))

	return b.socketfilepath
}

// PidFile returns path to the pid file of the brick process
func (b *Glusterfsd) PidFile() string {

	if b.pidfilepath != "" {
		return b.pidfilepath
	}

	rundir := config.GetString("rundir")
	brickPathWithoutSlashes := strings.Trim(strings.Replace(b.brickinfo.Path, "/", "-", -1), "-")
	// FIXME: The brick can no longer clean this up on clean shut down
	pidfilename := fmt.Sprintf("%s-%s.pid", b.brickinfo.NodeID.String(), brickPathWithoutSlashes)
	b.pidfilepath = path.Join(rundir, "gluster", pidfilename)

	return b.pidfilepath
}

// NewGlusterfsd returns a new instance of Glusterfsd type which implements the Daemon interface
func NewGlusterfsd(binfo Brickinfo) (*Glusterfsd, error) {
	path, e := exec.LookPath(glusterfsdBin)
	if e != nil {
		return nil, e
	}
	brickObject := &Glusterfsd{binarypath: path, brickinfo: binfo}
	return brickObject, nil
}

// ID returns the unique identifier of the brick. The brick path is unique
// on a node.
func (b *Glusterfsd) ID() string {
	return b.brickinfo.Path
}
