package georeplication

import (
	"fmt"
	"path"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"

	config "github.com/spf13/viper"
)

const (
	gsyncdCommand = "/usr/local/libexec/glusterfs/gsyncd"
)

// Gsyncd type represents information about Gsyncd process
type Gsyncd struct {
	// Externally consumable using methods of Gsyncd interface
	binarypath     string
	args           []string
	configFilePath string
	pidfilepath    string
	// For internal use
	sessioninfo georepapi.GeorepSession
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
func (g *Gsyncd) Args() []string {
	g.args = []string{}

	g.args = append(g.args, "monitor")
	g.args = append(g.args, g.sessioninfo.MasterVol)
	g.args = append(g.args,
		fmt.Sprintf("%s@%s::%s", g.sessioninfo.RemoteUser, g.sessioninfo.RemoteHosts[0].Hostname, g.sessioninfo.RemoteVol))
	g.args = append(g.args, "--local-node-id", gdctx.MyUUID.String())
	g.args = append(g.args, "-c", g.ConfigFile())
	g.args = append(g.args, "--use-gconf-volinfo")

	return g.args
}

// ConfigFile returns path to the config file
func (g *Gsyncd) ConfigFile() string {

	if g.configFilePath != "" {
		return g.configFilePath
	}

	g.configFilePath = path.Join(
		config.GetString("localstatedir"),
		"geo-replication",
		fmt.Sprintf("%s_%s_%s", g.sessioninfo.MasterVol, g.sessioninfo.RemoteHosts[0].Hostname, g.sessioninfo.RemoteVol),
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

	pidfilename := fmt.Sprintf("gsyncd-%s-%s-%s.pid", g.sessioninfo.MasterVol, g.sessioninfo.RemoteHosts[0].Hostname, g.sessioninfo.RemoteVol)
	g.pidfilepath = path.Join(config.GetString("rundir"), pidfilename)
	return g.pidfilepath
}

// newGsyncd returns a new instance of Gsyncd monitor type which implements the Daemon interface
func newGsyncd(sessioninfo georepapi.GeorepSession) (*Gsyncd, error) {
	return &Gsyncd{binarypath: gsyncdCommand, sessioninfo: sessioninfo}, nil
}

// ID returns the unique identifier of the gsyncd.
func (g *Gsyncd) ID() string {
	return g.sessioninfo.MasterID.String() + "-" + g.sessioninfo.RemoteID.String()
}

func (g *Gsyncd) statusArgs(localPath string) []string {
	return []string{
		"status",
		g.sessioninfo.MasterVol,
		fmt.Sprintf("%s@%s::%s", g.sessioninfo.RemoteUser, g.sessioninfo.RemoteHosts[0].Hostname, g.sessioninfo.RemoteVol),
		"-c",
		g.ConfigFile(),
		"--local-path",
		localPath,
		"--json"}
}
