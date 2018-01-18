package bitrot

import (
	"bytes"
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
	args           string
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
func (b *Bitd) Args() string {
	return b.args
}

// SocketFile returns path to the socket file
func (b *Bitd) SocketFile() string {
	if b.socketfilepath != "" {
		return b.socketfilepath
	}

	glusterdSockDir := path.Join(config.GetString("rundir"), "gluster")
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

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" -s %s", shost))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", b.volfileID))
	buffer.WriteString(fmt.Sprintf(" -p %s", b.pidfilepath))
	buffer.WriteString(fmt.Sprintf(" -l %s", b.logfile))
	buffer.WriteString(fmt.Sprintf(" -S %s", b.socketfilepath))
	buffer.WriteString(fmt.Sprintf(" --global-timer-wheel"))
	b.args = buffer.String()

	return b, nil
}

// ID returns the unique identifier of the bitd.
func (b *Bitd) ID() string {
	return ""
}
