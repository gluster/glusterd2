// Package rest implements the REST server for GlusterD
package rest

import (
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"gopkg.in/tylerb/graceful.v1"
)

// GDRest is the GlusterD Rest server
type GDRest struct {
	Routes *mux.Router
	srv    *graceful.Server
}

// New returns a GDRest object which can listen on the configured address
func New() *GDRest {
	rest := &GDRest{}

	rest.Routes = mux.NewRouter()

	n := negroni.New()
	n.UseHandler(rest.Routes)

	rest.srv = &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Handler: n,
		},
	}

	return rest
}

// Serve begins serving client HTTP requests served by REST server
func (r *GDRest) Serve(l net.Listener) error {
	go r.srv.Serve(l)
	log.WithField("ip:port", l.Addr().String()).Info("Started GlusterD REST server")
	return nil
}

// Stop stops the GlusterD Rest server
func (r *GDRest) Stop() {
	log.Debug("Stopping the GlusterD Rest server")
	schan := r.srv.StopChan()
	r.srv.Stop(10 * time.Second)
	<-schan
	log.Info("Stopped the GlusterD Rest Server")

	return
}
