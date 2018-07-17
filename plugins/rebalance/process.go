package rebalance

import (
	//        "bytes"
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
	args           []string
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
func (r *Process) Args() []string {

	volfileserver, port, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if volfileserver == "" {
		//	volfileserver = "127.0.0.1"
		volfileserver = "localhost"
	}

	volFileID := r.rInfo.Volname + "/rebalance"
	logDir := path.Join(config.GetString("logdir"), "glusterfs")
	logFile := fmt.Sprintf("%s/%s-rebalance.log", logDir, r.rInfo.Volname)
	cmd := r.rInfo.Cmd
	commithash := r.rInfo.CommitHash

	r.args = []string{}

	r.args = append(r.args, "-s", volfileserver)
	r.args = append(r.args, "--volfile-server-port", port)
	r.args = append(r.args, "--volfile-id", volFileID)
	r.args = append(r.args, "--process-name")
	r.args = append(r.args, "rebalance")
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("*distribute.use-readdirp=yes"))
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("*distribute.lookup-unhashed=yes"))
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("*distribute.assert-no-child-down=yes"))
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("replicate.data-self-heal=off"))
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("replicate.metadata-self-heal=off"))
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("replicate.entry-self-heal=off"))
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("*distribute.readdir-optimize=on"))
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("*distribute.rebalance-cmd=%d", cmd))
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("*distribute.node-uuid=%s", gdctx.MyUUID))
	r.args = append(r.args, "--xlator-option", fmt.Sprintf("*distribute.commit-hash=%d", commithash))
	r.args = append(r.args, "-p", r.PidFile())
	r.args = append(r.args, "--socket-file", r.SocketFile())
	r.args = append(r.args, "-l", logFile)

	return r.args
}

// SocketFile returns path to the socket file used for IPC
func (r *Process) SocketFile() string {

	if r.socketfilepath != "" {
		return r.socketfilepath
	}

	r.socketfilepath = path.Join(config.GetString("rundir"),
		fmt.Sprintf("%s-rebalance-%x.socket", r.rInfo.Volname, xxhash.Sum64String(gdctx.MyUUID.String())))
	return r.socketfilepath
}

// PidFile returns path to the pid file of rebalance process
func (r *Process) PidFile() string {
	if r.pidfilepath != "" {
		return r.pidfilepath
	}

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
	return r.rInfo.Volname + "-rebalance"
}
