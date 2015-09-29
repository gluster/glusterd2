// Package context is the runtime context of GlusterD
//
// Any package that needs access to the GlusterD runtime context just needs to
// import this package.
package context

import (
	"sync"

	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/store"
	"github.com/gluster/glusterd2/transaction"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

// Any object that is a part of the GlusterD context and needs to be available
// to other packages should be declared here as exported global variables
var (
	MyUUID uuid.UUID
	Rest   *rest.GDRest
	TxnFw  *transaction.GDTxnFw
	Store  *store.GDStore
)

var (
	initOnce sync.Once
)

func doInit() {
	log.Debug("Initializing GlusterD context")

	initLocalStateDir()

	initMyUUID()

	Rest = rest.New()

	Store = store.New()

	log.Debug("Initialized GlusterD context")
}

// Init initializes the GlusterD context. This should be called once before doing anything else.
func Init() {
	initOnce.Do(doInit)
}
