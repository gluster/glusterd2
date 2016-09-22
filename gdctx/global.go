// Package gdctx is the runtime context of GlusterD
//
// This file implements the global runtime context for GlusterD.
// Any package that needs access to the GlusterD global runtime context just
// needs to import this package.
package gdctx

import (
	"os"
	"sync"

	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	etcdclient "github.com/coreos/etcd/client"
	"github.com/pborman/uuid"
	config "github.com/spf13/viper"
)

// Various version constants that will be used by GD2
const (
	MaxOpVersion = 40000
	APIVersion   = 1
)

var (
	GlusterdVersion = "4.0-dev"
)

// Any object that is a part of the GlusterD context and needs to be available
// to other packages should be declared here as exported global variables
var (
	MyUUID         uuid.UUID
	Restart        bool // Indicates if its a fresh install or not
	Rest           *rest.GDRest
	OpVersion      int
	EtcdProcessCtx *os.Process
	EtcdClient     etcdclient.Client
	HostIP         string
)

var (
	initOnce sync.Once
)

func initOpVersion() {
	//TODO : Need cluster awareness and then decide the op-version
	OpVersion = MaxOpVersion
}

//initETCDClient will initialize etcd client that will be use during member add/remove in the cluster
func initETCDClient() error {
	SetLocalHostIP()
	c, err := etcdclient.New(etcdclient.Config{Endpoints: []string{"http://" + HostIP + ":2379"}})
	if err != nil {
		log.WithField("err", err).Error("Failed to create etcd client")
		return err
	}
	EtcdClient = c

	return nil
}

func doInit() {
	log.Debug("Initializing GlusterD context")

	utils.InitDir(config.GetString("localstatedir"))

	initOpVersion()

	Rest = rest.New()

	initStore()

	// Initializing etcd client
	err := initETCDClient()
	if err != nil {
		log.WithField("err", err).Error("Failed to initialize etcd client")
		return
	}

	log.Debug("Initialized GlusterD context")
}

// Init initializes the GlusterD context. This should be called once before doing anything else.
func Init() {
	initOnce.Do(doInit)
}

// GetEtcdMemberAPI returns the etcd MemberAPI
func GetEtcdMemberAPI() etcdclient.MembersAPI {
	var c etcdclient.Client
	return etcdclient.NewMembersAPI(c)
}

// AssignEtcdProcessCtx is to assign the etcd ctx in context.EtcdCtx
func AssignEtcdProcessCtx(ctx *os.Process) {
	EtcdProcessCtx = ctx
}

// SetLocalHostIP sets the local IP address
func SetLocalHostIP() {
	hostIP, err := utils.GetLocalIP()
	if err != nil {
		log.Fatal("Could not able to get IP address")
	}
	HostIP = hostIP
}
