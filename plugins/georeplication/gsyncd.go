package georeplication

import (
	"bytes"
	"fmt"
	"path"

	"github.com/gluster/glusterd2/gdctx"

	config "github.com/spf13/viper"
)

// Gsyncd type represents information about Gsyncd process
type Gsyncd struct {
	// Externally consumable using methods of Glusterfsd interface
	binarypath     string
	args           string
	configFilePath string
	pidfilepath    string
	// For internal use
	sessioninfo Session
}

// Name returns human-friendly name of the gsyncd process. This is used for logging.
func (g *Gsyncd) Name() string {
	return "gsyncd"
}

// Path returns absolute path to the binary of gsyncd process
func (g *Gsyncd) Path() string {
	return g.binarypath
}

// Args returns arguments to be passed to gsyncd process during spawn.
func (g *Gsyncd) Args() string {
	var buffer bytes.Buffer
	buffer.WriteString(" monitor")
	buffer.WriteString(fmt.Sprintf(" %s", g.sessioninfo.MasterVol))
	buffer.WriteString(fmt.Sprintf(" %s@%s::%s", g.sessioninfo.SlaveUser, g.sessioninfo.SlaveHosts[0], g.sessioninfo.SlaveVol))
	buffer.WriteString(fmt.Sprintf(" --local-node-id %s", gdctx.MyUUID.String()))
	buffer.WriteString(fmt.Sprintf(" -c %s", g.ConfigFile()))

	g.args = buffer.String()
	return g.args
}

func (g *Gsyncd) statusArgs(localPath string) string {
	var buffer bytes.Buffer
	buffer.WriteString(" status")
	buffer.WriteString(fmt.Sprintf(" %s", g.sessioninfo.MasterVol))
	buffer.WriteString(fmt.Sprintf(" %s@%s::%s", g.sessioninfo.SlaveUser, g.sessioninfo.SlaveHosts[0], g.sessioninfo.SlaveVol))
	buffer.WriteString(fmt.Sprintf(" -c %s", g.ConfigFile()))
	buffer.WriteString(fmt.Sprintf(" --local-path %s", localPath))

	return buffer.String()
}

// ConfigFile returns path to the config file
func (g *Gsyncd) ConfigFile() string {

	if g.configFilePath != "" {
		return g.configFilePath
	}

	g.configFilePath = path.Join(
		config.GetString("localstatedir"),
		"geo-replication",
		fmt.Sprintf("%s_%s_%s", g.sessioninfo.MasterVol, g.sessioninfo.SlaveHosts[0], g.sessioninfo.SlaveVol),
		"gsyncd.conf",
	)
	return g.configFilePath
}

// SocketFile returns path to the socket file, Gsyncd doesn't have socket file. This func is required for Daemon interface
func (g *Gsyncd) SocketFile() string {
	return ""
}

// PidFile returns path to the pid file of the gsyncd monitor process
func (g *Gsyncd) PidFile() string {

	if g.pidfilepath != "" {
		return g.pidfilepath
	}

	rundir := config.GetString("rundir")
	pidfilename := fmt.Sprintf("gsyncd-%s-%s-%s.pid", g.sessioninfo.MasterVol, g.sessioninfo.SlaveHosts[0], g.sessioninfo.SlaveVol)
	g.pidfilepath = path.Join(rundir, "gluster", pidfilename)
	return g.pidfilepath
}

// NewGsyncd returns a new instance of Gsyncd monitor type which implements the Daemon interface
func NewGsyncd(sessioninfo Session) (*Gsyncd, error) {
	// TODO Change this path to dynamic
	path := "/usr/local/libexec/glusterfs/gsyncd"
	sessionObject := &Gsyncd{binarypath: path, sessioninfo: sessioninfo}
	return sessionObject, nil
}

// ID returns the unique identifier of the gsyncd.
func (g *Gsyncd) ID() string {
	return g.sessioninfo.MasterID.String() + "-" + g.sessioninfo.SlaveID.String()
}
