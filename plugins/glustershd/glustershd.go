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
	binarypath     string
	args           string
	socketfilepath string
	pidfilepath    string
}

// Name returns human-friendly name of the glustershd process. This is used for
// logging
func (shd *Glustershd) Name() string {
	return "glustershd"
}

// Path returns absolute path of the binary of glustershd process
func (shd *Glustershd) Path() string {
	return shd.binarypath
}

// Args returns arguments to be passed to glustershd process during spawn.
func (shd *Glustershd) Args() string {
	shost, _, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if shost == "" {
		shost = "localhost"
	}
	volFileID := "gluster/glustershd"

	logFile := path.Join(config.GetString("logdir"), "glusterfs", "glustershd.log")
	glusterdSockDir := path.Join(config.GetString("rundir"), "gluster")
	socketfilepath := fmt.Sprintf("%s/%x.socket", glusterdSockDir, xxhash.Sum64String(gdctx.MyUUID.String()))

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" -s %s", shost))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", volFileID))
	buffer.WriteString(fmt.Sprintf(" -p %s", shd.PidFile()))
	buffer.WriteString(fmt.Sprintf(" -l %s", logFile))
	buffer.WriteString(fmt.Sprintf(" -S %s", socketfilepath))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *replicate*.node-uuid=%s", gdctx.MyUUID))

	shd.args = buffer.String()
	return shd.args
}

// SocketFile returns path to the socket file used for IPC.
func (shd *Glustershd) SocketFile() string {
	return ""
}

// PidFile returns path to the pid file of self heal process.
func (shd *Glustershd) PidFile() string {

	if shd.pidfilepath != "" {
		return shd.pidfilepath
	}

	rundir := config.GetString("rundir")
	shd.pidfilepath = path.Join(rundir, "gluster", "glustershd.pid")

	return shd.pidfilepath
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

// ID returns the uniques identifier for the daemon. One daemon is required per node
// that's why glustershd will be sufficient.
func (shd *Glustershd) ID() string {
	return "glustershd"
}
