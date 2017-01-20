// Package rest implements the REST server for GlusterD
package rest

import (
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/soheilhy/cmux"
	"gopkg.in/tylerb/graceful.v1"
)

// GDRest is the GlusterD Rest server
type GDRest struct {
	Routes   *mux.Router
	server   *graceful.Server
	listener net.Listener
}

// New returns a GDRest object which can listen on the configured address
func New(l net.Listener) *GDRest {
	rest := &GDRest{}

	rest.Routes = mux.NewRouter()

	n := negroni.New()
	n.UseHandler(rest.Routes)

	rest.server = &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Handler: n,
		},
	}

	rest.listener = l

	return rest
}

// NewMuxed returns a GDRest object which listens on a CMux multiplexed connection
func NewMuxed(m cmux.CMux) *GDRest {
	return New(m.Match(cmux.HTTP1Fast()))
}

// Serve begins serving client HTTP requests served by REST server
func (r *GDRest) Serve() {
	r.registerRoutes()
	log.WithField("ip:port", r.listener.Addr().String()).Info("Started GlusterD REST server")
	r.server.Serve(r.listener)
	return
}

// Stop stops the GlusterD Rest server
func (r *GDRest) Stop() {
	log.Debug("Stopping the GlusterD Rest server")
	schan := r.server.StopChan()
	r.server.Stop(10 * time.Second)
	<-schan
	log.Info("Stopped the GlusterD Rest Server")

	return
}
