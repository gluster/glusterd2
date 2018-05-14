package rebalance

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"path"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"

	"github.com/cespare/xxhash"
	config "github.com/spf13/viper"
)

const (
	glusterfsBin = "glusterfs"
)

// Process type represents information about rebalance process
type Process struct {
	binarypath     string
	args           string
	socketfilepath string
	pidfilepath    string
	rInfo          rebalanceapi.RebalInfo
}

// Name returns the process name
func (r *Process) Name() string {
	return "rebalance"
}

// Path returns absolute path to the binary of rebalance process
func (r *Process) Path() string {
	return r.binarypath
}

// Args returns arguments to be passed to rebalance process
func (r *Process) Args() string {

	volfileserver, port, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if volfileserver == "" {
		volfileserver = "localhost"
	}

	volFileID := path.Join("rebalance/", r.rInfo.Volname)
	logDir := path.Join(config.GetString("logdir"), "glusterfs")
	logFile := fmt.Sprintf("%s/%s-rebalance.log", logDir, r.rInfo.Volname)
	cmd := r.rInfo.Cmd
	commithash := r.rInfo.CommitHash

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" -s %s --volfile-server-port %s", volfileserver, port))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", volFileID))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.use-readdirp=yes"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.lookup-unhashed=yes"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.assert-no-child-down=yes"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option replicate.data-self-heal=off"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option replicate.metadata-self-heal=off"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option replicate.entry-self-heal=off"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *distribute.readdir-optimize=on"))
	buffer.WriteString(fmt.Sprintf(" --process-name rebalance"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *distribute.rebalance-cmd=%d", cmd))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *distribute.node-uuid=%s", gdctx.MyUUID))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *distribute.commit-hash=%d", commithash))
	buffer.WriteString(fmt.Sprintf(" -p %s", r.PidFile()))
	buffer.WriteString(fmt.Sprintf(" --socket-file %s", r.SocketFile()))
	buffer.WriteString(fmt.Sprintf(" -l %s", logFile))
	r.args = buffer.String()
	return r.args
}

// SocketFile returns path to the socket file used for IPC
func (r *Process) SocketFile() string {

	if r.socketfilepath != "" {
		return r.socketfilepath
	}

	glusterdSockDir := path.Join(config.GetString("rundir"), "", r.rInfo.Volname)
	r.socketfilepath = fmt.Sprintf("%s-rebalance-%x.socket", glusterdSockDir, xxhash.Sum64String(gdctx.MyUUID.String()))
	return r.socketfilepath
}

// PidFile returns path to the pid file of rebalance process
func (r *Process) PidFile() string {
	if r.pidfilepath != "" {
		return r.pidfilepath
	}
//	pidDir := path.Join(config.GetString("rundir"), "gluster")

	r.pidfilepath = fmt.Sprintf("%s/%s-rebalance.pid", config.GetString("rundir"), r.rInfo.Volname)
	return r.pidfilepath
}

// NewRebalanceProcess returns a new instance of Glusterfsd type which implements the Daemon interface
func NewRebalanceProcess(rinfo rebalanceapi.RebalInfo) (*Process, error) {
	path, e := exec.LookPath(glusterfsBin)
	if e != nil {
		return nil, e
	}
	rebalanceObject := &Process{binarypath: path, rInfo: rinfo}
	return rebalanceObject, nil
}

// ID returns the unique identifier on a node
func (r *Process) ID() string {
	return "rebalance-" + r.rInfo.Volname
}
