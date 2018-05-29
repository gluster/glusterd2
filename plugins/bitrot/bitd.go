package bitrot

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"

	"github.com/cespare/xxhash"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	config "github.com/spf13/viper"
)

const (
	bitdBin = "glusterfs"
)

// Bitd type represents information about bitd process
type Bitd struct {
	args           []string
	pidfilepath    string
	binarypath     string
	volfileID      string
	logfile        string
	socketfilepath string
}

// Name returns human-friendly name of the bitd process. This is used for logging.
func (b *Bitd) Name() string {
	return "bitd"
}

// Path returns absolute path to the binary of bitd process
func (b *Bitd) Path() string {
	return b.binarypath
}

// Args returns arguments to be passed to bitd process during spawn.
func (b *Bitd) Args() []string {
	return b.args
}

// SocketFile returns path to the socket file
func (b *Bitd) SocketFile() string {
	if b.socketfilepath != "" {
		return b.socketfilepath
	}

	glusterdSockDir := config.GetString("rundir")
	b.socketfilepath = fmt.Sprintf("%s/bitd-%x.socket", glusterdSockDir, xxhash.Sum64String(gdctx.MyUUID.String()))

	return b.socketfilepath
}

// PidFile returns path to the pid file of the bitd process
func (b *Bitd) PidFile() string {
	return b.pidfilepath
}

// newBitd returns a new instance of bitd type which implements the Daemon interface
func newBitd() (*Bitd, error) {
	binarypath, e := exec.LookPath(bitdBin)
	if e != nil {
		return nil, e
	}

	b := &Bitd{binarypath: binarypath}
	b.volfileID = gdctx.MyUUID.String() + "-gluster/bitd"
	b.logfile = path.Join(config.GetString("logdir"), "glusterfs", "bitd.log")

	// Create pidFiledir dir
	pidFileDir := fmt.Sprintf("%s/bitd", config.GetString("rundir"))
	e = os.MkdirAll(pidFileDir, os.ModeDir|os.ModePerm)
	if e != nil {
		return nil, e
	}
	b.pidfilepath = fmt.Sprintf("%s/bitd.pid", pidFileDir)
	b.socketfilepath = b.SocketFile()

	shost, _, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if shost == "" {
		shost = "localhost"
	}

	b.args = []string{}
	b.args = append(b.args, "-s", shost)
	b.args = append(b.args, "--volfile-id", b.volfileID)
	b.args = append(b.args, "-p", b.pidfilepath)
	b.args = append(b.args, "-l", b.logfile)
	b.args = append(b.args, "-S", b.socketfilepath)
	b.args = append(b.args, "--global-timer-wheel")

	return b, nil
}

// ID returns the unique identifier of the bitd.
func (b *Bitd) ID() string {
	return ""
}
