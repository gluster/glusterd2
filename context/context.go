// Package context is the runtime context of GlusterD
//
// Any package that needs access to the GlusterD runtime context just needs to
// import this package.
package context

import (
	"os"
	"sync"

	"github.com/gluster/glusterd2/config"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

// Various version constants that will be used by GD2
const (
	MaxOpVersion    = 40000
	APIVersion      = 1
	GlusterdVersion = "4.0-dev"
)

// Any object that is a part of the GlusterD context and needs to be available
// to other packages should be declared here as exported global variables
var (
	MyUUID    uuid.UUID
	Rest      *rest.GDRest
	TxnFw     *transaction.GDTxnFw
	OpVersion int
	EtcdCtx   *os.Process
)

var (
	initOnce sync.Once
)

func initOpVersion() {
	//TODO : Need cluster awareness and then decide the op-version
	OpVersion = MaxOpVersion
}

func doInit() {
	log.Debug("Initializing GlusterD context")

	utils.InitDir(config.LocalStateDir)

	initMyUUID()
	initOpVersion()

	Rest = rest.New()

	initStore()

	log.Debug("Initialized GlusterD context")
}

// Init initializes the GlusterD context. This should be called once before doing anything else.
func Init() {
	initOnce.Do(doInit)
}

// AssignEtcdCtx () is to assign the etcd ctx in context.EtcdCtx
func AssignEtcdCtx(ctx *os.Process) {
	EtcdCtx = ctx
}
