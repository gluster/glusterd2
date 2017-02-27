package mgmt

import (
	"path"

	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	mgmt "github.com/purpleidea/mgmt/lib"
	config "github.com/spf13/viper"
)

// Mgmt implements the supervisor interface
type Mgmt struct {
	m *mgmt.Main
}

// New returns a newly initialized Mgmt. Returns nil if it couldn't init.
func New() *Mgmt {
	m := &mgmt.Main{}
	m.Program = "glusterd2"

	prefix := path.Join(config.GetString("localstatedir"), "mgmt")
	utils.InitDir(prefix)

	m.Prefix = &prefix

	// TODO: Set all of this correctly.
	m.IdealClusterSize = -1
	m.ConvergedTimeout = -1
	m.Noop = false

	m.GAPI = &GlusterGAPI{
		Name: "glusterd2",
	}

	if e := m.Init(); e != nil {
		log.WithError(e).Error("failed to initialize mgmt")
		return nil
	}
	return &Mgmt{m}
}

// Serve starts the mgmt engine
func (m *Mgmt) Serve() {
	if e := m.m.Run(); e != nil {
		log.WithError(e).Error("failed to run mgmt")
	}
}

// Stop stops the mgmt engine
func (m *Mgmt) Stop() {
	m.m.Exit(nil)
}
