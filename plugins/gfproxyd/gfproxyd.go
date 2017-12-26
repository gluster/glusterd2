package gfproxyd

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"path"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/pmap"
	config "github.com/spf13/viper"
)

const (
	gfproxydBin = "glusterfsd"
)

// gfproxyd type represents information about gfproxyd process
type gfproxyd struct {
	volname      string
	args         string
	pidfilepath  string
	binarypath   string
	volfileID    string
	logfile      string
	gfproxyID    string
	gfproxydPort string
}

// Name returns human-friendly name of the gfproxyd process. This is used for logging.
func (g *gfproxyd) Name() string {
	return g.gfproxyID
}

// Path returns absolute path to the binary of gfproxyd process
func (g *gfproxyd) Path() string {
	return g.binarypath
}

// Args returns arguments to be passed to gfproxyd process during spawn.
func (g *gfproxyd) Args() string {
	return g.args
}

// SocketFile returns path to the socket file
func (g *gfproxyd) SocketFile() string {
	return ""
}

// PidFile returns path to the pid file of the gfproxyd process
func (g *gfproxyd) PidFile() string {
	return g.pidfilepath
}

// newgfproxyd returns a new instance of gfproxyd type which implements the Daemon interface
func newgfproxyd(volname string) (*gfproxyd, error) {
	g := &gfproxyd{volname: volname}
	binarypath, e := exec.LookPath(gfproxydBin)
	if e != nil {
		return nil, e
	}
	g.binarypath = binarypath
	g.gfproxyID = fmt.Sprintf("gfproxyd-%s", volname)
	g.gfproxydPort = strconv.Itoa(pmap.AssignPort(0, g.gfproxyID))
	g.volfileID = fmt.Sprintf("gfproxyd/%s", volname)
	g.logfile = path.Join(config.GetString("logdir"), "glusterfs", fmt.Sprintf("gfproxyd-%s.log", volname))
	g.pidfilepath = fmt.Sprintf("%s/gfproxyd-%s.pid", config.GetString("rundir"), volname)

	shost, _, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if shost == "" {
		shost = "localhost"
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" -s %s", shost))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", g.volfileID))
	buffer.WriteString(fmt.Sprintf(" -p %s", g.pidfilepath))
	buffer.WriteString(fmt.Sprintf(" -l %s", g.logfile))
	buffer.WriteString(fmt.Sprintf(" --brick-name %s", g.gfproxyID))
	buffer.WriteString(fmt.Sprintf(" --brick-port %s", g.gfproxydPort))
	buffer.WriteString(fmt.Sprintf(" --xlator-option %s-server.listen-port=%s", g.volname, g.gfproxydPort))

	g.args = buffer.String()

	return g, nil
}

// ID returns the unique identifier of the gfproxyd.
func (g *gfproxyd) ID() string {
	return g.gfproxyID
}
