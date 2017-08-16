// Package rest implements the REST server for GlusterD
package rest

import (
	"net"
	"net/http"

	"github.com/gluster/glusterd2/middleware"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/soheilhy/cmux"
)

// GDRest is the GlusterD Rest server
type GDRest struct {
	Routes   *mux.Router
	listener net.Listener
}

// New returns a GDRest object which can listen on the configured address
func New(l net.Listener) *GDRest {
	rest := &GDRest{
		mux.NewRouter(),
		l,
	}

	rest.registerRoutes()

	return rest
}

// NewMuxed returns a GDRest object which listens on a CMux multiplexed connection
func NewMuxed(m cmux.CMux) *GDRest {
	return New(m.Match(cmux.HTTP1Fast()))
}

// Serve begins serving client HTTP requests served by REST server
func (r *GDRest) Serve() {
	chain := alice.New(middleware.Auth, middleware.LogRequest, middleware.ReqIDGenerator).Then(r.Routes)
	log.WithField("ip:port", r.listener.Addr().String()).Info("Started GlusterD ReST server")
	if err := http.Serve(r.listener, chain); err != nil {
		//TODO: Correctly handle valid errors. We could also be having errors when stopping
		log.WithError(err).Error("GlusterD ReST server failed")
	}
	return
}

// Stop stops the GlusterD Rest server
func (r *GDRest) Stop() {
	log.Debug("stopping the GlusterD ReST server")
	// TODO: Graceful shutdown here
	log.Info("stopped GlusterD ReST server")
}
