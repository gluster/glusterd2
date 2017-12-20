package rebalance

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
	sbinDir = "glusterfs"
)

// RebalanceProcess type represents information required to start rebalance process
type RebalanceProcess struct {
	binarypath     string
	args           string
	socketfilepath string
	pidfilepath    string
	rebalanceinfo  RebalanceInfo
}

// Name returns the process name
func (r *RebalanceProcess) Name() string {
	return "rebalance"
}

// Path returns absolute path to the binary of rebalance process
func (r *RebalanceProcess) Path() string {
	return r.binarypath
}

// Args returns arguments to be passed to rebalance process
func (r *RebalanceProcess) Args() string {

	shost, _, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if shost == "" {
		shost = "localhost"
	}

	volFileID := path.Join("rebalance/", r.rebalanceinfo.Volname)
	File := path.Join(config.GetString("logdir"), "glusterfs", r.rebalanceinfo.Volname)
	logFile := fmt.Sprintf("%s-rebalance.log", File)
	cmd := r.rebalanceinfo.Status
	commithash := r.rebalanceinfo.CommitHash

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" -s %s", shost))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", volFileID))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.use-readdirp=yes"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.lookup-unhashed=yes"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.assert-no-child-down=yes"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option replicate.data-self-heal=off"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option replicate.metadata-self-heal=off"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option replicate.entry-self-heal=off"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.readdir-optimize=on"))
	buffer.WriteString(fmt.Sprintf(" --process-name rebalance"))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.rebalance-cmd=%d", cmd))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.node-uuid=%s", gdctx.MyUUID))
	buffer.WriteString(fmt.Sprintf(" --xlator-option *dht.commit-hash=%d", commithash))
	buffer.WriteString(fmt.Sprintf(" -p %s", r.PidFile()))
	buffer.WriteString(fmt.Sprintf(" -S %s", r.SocketFile()))
	buffer.WriteString(fmt.Sprintf(" -l %s", logFile))
	r.args = buffer.String()
	return r.args
}

// SocketFile returns path to the socket file used for IPC
func (r *RebalanceProcess) SocketFile() string {

	if r.socketfilepath != "" {
		return r.socketfilepath
	}

	glusterdSockDir := path.Join(config.GetString("rundir"), "gluster", "gluster-rebalance")
	r.socketfilepath = fmt.Sprintf("%s-%x.socket", glusterdSockDir, xxhash.Sum64String(gdctx.MyUUID.String()))
	return r.socketfilepath
}

// PidFile returns path to the pid file of rebalance process
func (r *RebalanceProcess) PidFile() string {
	if r.pidfilepath != "" {
		return r.pidfilepath
	}
	filepath := path.Join(config.GetString("rundir"), "gluster", "rebalance")
	r.pidfilepath = fmt.Sprintf("%s.pid", filepath)
	return r.pidfilepath
}

// NewRebalanceProcess returns a new instance of Glusterfsd type which implements the Daemon interface
func NewRebalanceProcess(rinfo RebalanceInfo) (*RebalanceProcess, error) {
	path, e := exec.LookPath(sbinDir)
	if e != nil {
		return nil, e
	}
	rebalanceObject := &RebalanceProcess{binarypath: path, rebalanceinfo: rinfo}
	return rebalanceObject, nil
}

// ID returns the unique identifier on a node
func (r *RebalanceProcess) ID() string {
	return "rebalance"
}
