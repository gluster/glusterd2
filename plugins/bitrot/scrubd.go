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
	scrubdBin = "glusterfs"
)

// Scrubd type represents information about scrubd process
type Scrubd struct {
	args           string
	pidfilepath    string
	binarypath     string
	volfileID      string
	logfile        string
	socketfilepath string
}

// Name returns human-friendly name of the scrubd process. This is used for logging.
func (s *Scrubd) Name() string {
	return "scrubd"
}

// Path returns absolute path to the binary of scrubd process
func (s *Scrubd) Path() string {
	return s.binarypath
}

// Args returns arguments to be passed to scrubd process during spawn.
func (s *Scrubd) Args() string {
	return s.args
}

// SocketFile returns path to the socket file
func (s *Scrubd) SocketFile() string {
	if s.socketfilepath != "" {
		return s.socketfilepath
	}

	glusterdSockDir := config.GetString("rundir")
	s.socketfilepath = fmt.Sprintf("%s/scrub-%x.socket", glusterdSockDir, xxhash.Sum64String(gdctx.MyUUID.String()))

	return s.socketfilepath
}

// PidFile returns path to the pid file of the scrubd process
func (s *Scrubd) PidFile() string {
	return s.pidfilepath
}

// newScrubd returns a new instance of scrubd type which implements the Daemon interface
func newScrubd() (*Scrubd, error) {
	binarypath, e := exec.LookPath(scrubdBin)
	if e != nil {
		return nil, e
	}

	s := &Scrubd{binarypath: binarypath}
	s.volfileID = gdctx.MyUUID.String() + "-gluster/scrub"
	s.logfile = path.Join(config.GetString("logdir"), "glusterfs", "scrub.log")

	// Create pidFiledir dir
	pidFileDir := fmt.Sprintf("%s/scrub", config.GetString("rundir"))
	e = os.MkdirAll(pidFileDir, os.ModeDir|os.ModePerm)
	if e != nil {
		return nil, e
	}
	s.pidfilepath = fmt.Sprintf("%s/scrub.pid", pidFileDir)
	s.socketfilepath = s.SocketFile()

	shost, _, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if shost == "" {
		shost = "localhost"
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" -s %s", shost))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", s.volfileID))
	buffer.WriteString(fmt.Sprintf(" -p %s", s.pidfilepath))
	buffer.WriteString(fmt.Sprintf(" -l %s", s.logfile))
	buffer.WriteString(fmt.Sprintf(" -S %s", s.socketfilepath))
	buffer.WriteString(fmt.Sprintf(" --global-timer-wheel"))
	s.args = buffer.String()

	return s, nil
}

// ID returns the unique identifier of the scrubd.
func (s *Scrubd) ID() string {
	return ""
}
