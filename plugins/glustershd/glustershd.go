package glustershd

import (
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
	args           []string
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
func (shd *Glustershd) Args() []string {
	shost, _, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if shost == "" {
		shost = "localhost"
	}
	volFileID := "gluster/glustershd"

	logFile := path.Join(config.GetString("logdir"), "glusterfs", "glustershd.log")

	shd.args = []string{}
	shd.args = append(shd.args, "-s", shost)
	shd.args = append(shd.args, "--volfile-id", volFileID)
	shd.args = append(shd.args, "-p", shd.PidFile())
	shd.args = append(shd.args, "-l", logFile)
	shd.args = append(shd.args, "-S", shd.SocketFile())
	shd.args = append(shd.args,
		"--xlator-option",
		fmt.Sprintf("*replicate*.node-uuid=%s", gdctx.MyUUID))

	return shd.args
}

// SocketFile returns path to the socket file used for IPC.
func (shd *Glustershd) SocketFile() string {
	if shd.socketfilepath != "" {
		return shd.socketfilepath
	}
	shd.socketfilepath = fmt.Sprintf("%s/shd-%x.socket", config.GetString("rundir"), xxhash.Sum64String(gdctx.MyUUID.String()))

	return shd.socketfilepath
}

// PidFile returns path to the pid file of self heal process.
func (shd *Glustershd) PidFile() string {

	if shd.pidfilepath != "" {
		return shd.pidfilepath
	}

	shd.pidfilepath = path.Join(config.GetString("rundir"), "glustershd.pid")

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
