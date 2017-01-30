// Package rest implements the REST server for GlusterD
package rest

import (
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
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
	log.WithField("ip:port", r.listener.Addr().String()).Info("Started GlusterD REST server")
	http.Serve(r.listener, r.Routes)
	return
}

// Stop stops the GlusterD Rest server
func (r *GDRest) Stop() {
	log.Debug("Stopping the GlusterD Rest server")
	r.listener.Close()
	// TODO: Graceful shutdown here
	return
}
