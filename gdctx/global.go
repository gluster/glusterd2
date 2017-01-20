// Package gdctx is the runtime context of GlusterD
//
// This file implements the global runtime context for GlusterD.
// Any package that needs access to the GlusterD global runtime context just
// needs to import this package.
package gdctx

import (
	"os"
	"sync"

	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
	config "github.com/spf13/viper"
)

// Various version constants that will be used by GD2
const (
	MaxOpVersion = 40000
	APIVersion   = 1
)

var (
	// GlusterdVersion is the version of the glusterd daemon
	GlusterdVersion = "4.0-dev"
)

// Any object that is a part of the GlusterD context and needs to be available
// to other packages should be declared here as exported global variables
var (
	MyUUID    uuid.UUID
	Restart   bool // Indicates if its a fresh install or not
	OpVersion int
	HostIP    string
	HostName  string
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

	utils.InitDir(config.GetString("localstatedir"))

	initOpVersion()

	// When glusterd is started for the first time, we will have Restart set to
	// false. That is when we'll have to initialize prefixes by passing true to
	// InitStore(). On subsequent restarts of glusterd, we would want to skip
	// initializing prefixes by passing false to InitStore()
	InitStore(!Restart)

	log.Debug("Initialized GlusterD context")
}

// Init initializes the GlusterD context. This should be called once before doing anything else.
func Init() {
	initOnce.Do(doInit)
}

// SetHostnameAndIP sets the local IP address and host name
func SetHostnameAndIP() {
	hostIP, err := utils.GetLocalIP()
	if err != nil {
		log.Fatal("SetHostnameAndIP: Could not get IP address")
	}
	HostIP = hostIP

	hostName, err := os.Hostname()
	if err != nil {
		log.Fatal("SetHostnameAndIP: Could not get hostname")
	}
	HostName = hostName
}
