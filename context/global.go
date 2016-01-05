// Package context is the runtime context of GlusterD
//
// This file implements the global runtime context for GlusterD.
// Any package that needs access to the GlusterD global runtime context just
// needs to import this package.
package context

import (
	"sync"

	"github.com/gluster/glusterd2/rest"

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
	OpVersion int
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

	initLocalStateDir()

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
