package quota

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
	quotadBin = "glusterfs"
)

// Quotad type represents information about the quota daemon
type Quotad struct {
	// Externally consumable using methods of Quotad interface
	binarypath     string
	args           []string
	socketfilepath string
	pidfilepath    string
	logfilepath    string
	volname        string
	volFileID      string

	// For internal use
}

// Name returns human-friendly name of the quota process. This is used for logging.
func (q *Quotad) Name() string {
	return "quotad"
}

// Path returns absolute path to the binary of quota process
func (q *Quotad) Path() string {
	return q.binarypath
}

// Args returns arguments to be passed to quota process during spawn.
func (q *Quotad) Args() []string {

	shost, _, _ := net.SplitHostPort(config.GetString("clientaddress"))
	if shost == "" {
		shost = "127.0.0.1"
	}
	q.volFileID = "gluster/quotad"

	q.args = []string{}
	q.args = append(q.args, "-s", shost)
	q.args = append(q.args, "--volfile-id", q.volFileID)
	q.args = append(q.args, "-p", q.PidFile())
	q.args = append(q.args, "-S", q.SocketFile())
	q.args = append(q.args, "-l", q.logfilepath)
	q.args = append(q.args, "--process-name", "quotad")
	q.args = append(q.args, "--xlator-option", "*replicate*.entry-self-heal=off")
	q.args = append(q.args, "--xlator-option", "*replicate*.metadata-self-heal=off")
	q.args = append(q.args, "--xlator-option", "replicate*.data-self-heal=off")

	return q.args
}

// SocketFile returns path to the quota socket file used for IPC.
func (q *Quotad) SocketFile() string {

	if q.socketfilepath != "" {
		return q.socketfilepath
	}

	glusterdSockDir := config.GetString("rundir")
	q.socketfilepath = fmt.Sprintf("%s/quotad-%x.socket", glusterdSockDir,
		xxhash.Sum64String(gdctx.MyUUID.String()))

	return q.socketfilepath
}

// PidFile returns path to the pid file of the quota process
func (q *Quotad) PidFile() string {
	return q.pidfilepath
}

// NewQuotad returns a new instance of Quotad type which implements the Daemon interface
func NewQuotad() (*Quotad, error) {
	binarypath, e := exec.LookPath(quotadBin)
	if e != nil {
		return nil, e
	}
	pidfilepath := path.Join(
		config.GetString("rundir"), "quotad.pid")
	logfilepath := path.Join(
		config.GetString("logdir"), "glusterfs",
		"quotad", "quotad.log",
	)

	return &Quotad{binarypath: binarypath, pidfilepath: pidfilepath,
		logfilepath: logfilepath}, nil
}

// ID returns the unique identifier of the quota.
func (q *Quotad) ID() string {
	return "quotad"
}
