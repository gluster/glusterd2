// Package context is the runtime context of GlusterD
//
// Any package that needs access to the GlusterD runtime context just needs to
// import this package.
package context

import (
	"sync"

	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/store"
	"github.com/kshlm/glusterd2/transaction"

	log "github.com/Sirupsen/logrus"
)

// Any object that is a part of the GlusterD context and needs to be available
// to other packages should be declared here as exported global variables
var (
	Rest  *rest.GDRest
	TxnFw *transaction.GDTxnFw
	Store *store.GDStore
)

var (
	initOnce sync.Once
)

func doInit() {
	log.Debug("Initializing GlusterD context")

	Rest = rest.New()

	Store = store.New()

	log.Debug("Initialized GlusterD context")
}

// Init initializes the GlusterD context. This should be called once before doing anything else.
func Init() {
	initOnce.Do(doInit)
}
