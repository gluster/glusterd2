package glustershd

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"path"

	"github.com/cespare/xxhash"
	"github.com/gluster/glusterd2/glusterd2/gdctx"

	config "github.com/spf13/viper"
)

const (
	glustershdBin = "glusterfs"
)

// Glustershd type represents information about Glustershd process
type Glustershd struct {
	//Externally consumable using methods of Gsyncd interface
	binarypath     string
	args           string
	socketfilepath string
	pidfilepath    string
}

// Name returns human-friendly name of the glustershd process. This is used for
// logging
func (b *Glustershd) Name() string {
	return "glustershd"
}

//Path returns absolute path of the binary of glustershd process
func (b *Glustershd) Path() string {
	return b.binarypath
}

// Args returns arguments to be passed to glustershd process during spawn.
func (b *Glustershd) Args() string {
	shost, sport, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if shost == "" {
		shost = "localhost"
	}
	_ = sport
	volFileID := "gluster/glustershd"

	logFile := path.Join(config.GetString("logdir"), "glusterfs", "glustershd.log")

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" -s %s", shost))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", volFileID))
	buffer.WriteString(fmt.Sprintf(" -p %s", b.PidFile()))
	buffer.WriteString(fmt.Sprintf(" -l %s", logFile))
	buffer.WriteString(fmt.Sprintf(" -S %s", b.SocketFile()))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *replicate*.node-uuid=%s", gdctx.MyUUID))

	b.args = buffer.String()
	return b.args
}

// SocketFile returns path to the socket file used for IPC.
func (b *Glustershd) SocketFile() string {

	if b.socketfilepath != "" {
		return b.socketfilepath
	}

	glusterdSockDir := path.Join(config.GetString("rundir"), "gluster")
	b.socketfilepath = fmt.Sprintf("%s/%x.socket", glusterdSockDir, xxhash.Sum64String(gdctx.MyUUID.String()))

	return b.socketfilepath
}

// PidFile returns path to the pid file of self heal process.
func (b *Glustershd) PidFile() string {

	if b.pidfilepath != "" {
		return b.pidfilepath
	}

	rundir := config.GetString("rundir")
	b.pidfilepath = path.Join(rundir, "gluster", "glustershd.pid")

	return b.pidfilepath
}

// newGlustershd returns a new instance of Glustershd type which implements the
// daemon interface
func newGlustershd() (*Glustershd, error) {
	path, e := exec.LookPath(glustershdBin)
	if e != nil {
		return nil, e
	}
	glustershdObject := &Glustershd{binarypath: path}
	return glustershdObject, nil
}

// ID returns the uniques identifier of the brick. The brick path is unique on a
// node
func (b *Glustershd) ID() string {
	return "glustershd"
}
