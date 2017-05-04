package mgmt

import (
	"path"

	"github.com/gluster/glusterd2/mgmt/gapi"

	log "github.com/Sirupsen/logrus"
	mgmt "github.com/purpleidea/mgmt/lib"
	config "github.com/spf13/viper"
)

type Mgmt struct {
	*mgmt.Main
}

func New() *Mgmt {
	// set all the options we want here...
	libmgmt := &mgmt.Main{}
	libmgmt.Program = "glusterd2"
	libmgmt.Version = "testing" // TODO: set on compilation
	p := path.Join(config.GetString("localstatedir"), "mgmt")
	libmgmt.Prefix = &p // enable for easy debugging
	libmgmt.IdealClusterSize = -1
	libmgmt.ConvergedTimeout = -1
	libmgmt.Noop = false // FIXME: careful!
	libmgmt.NoPgp = true
	libmgmt.Seeds = []string{"http://" + config.GetString("etcdclientaddress")}
	libmgmt.NoServer = true

	libmgmt.GAPI = &gapi.Gd3GAPI{ // graph API
		Program: "gd2",
		Version: "testing",
	}

	if err := libmgmt.Init(); err != nil {
		log.WithError(err).Fatal("Init failed")
	}
	return &Mgmt{libmgmt}
}

func (m *Mgmt) Serve() {
	m.Main.Run()
}

func (m *Mgmt) Stop() {
	m.Main.Exit(nil)
}
