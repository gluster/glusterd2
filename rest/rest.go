// Package rest implements the REST server for GlusterD
package rest

import (
	"net/http"
	"time"

	"github.com/kshlm/glusterd2/config"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"gopkg.in/tylerb/graceful.v1"
)

type GDRest struct {
	Routes *mux.Router

	srv *graceful.Server
}

// function New returns a GlusterDRest object which listens on the give address
func New() *GDRest {
	rest := &GDRest{}

	rest.Routes = mux.NewRouter()

	n := negroni.New()
	n.Use(&negronilogrus.Middleware{Logger: log.StandardLogger(), Name: "glusterd-rest"})
	n.UseHandler(rest.Routes)

	rest.srv = &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:    *config.RestAddress,
			Handler: n,
		},
	}

	return rest
}

func (r *GDRest) Listen() error {
	log.Debug("Beginning the GlusterD Rest server")
	err := r.srv.ListenAndServe()
	if err != nil {
		log.WithField("error", err).Error("Failed to start the Rest Server")
		return err
	}
	return nil
}

func (r *GDRest) Stop() {
	log.Debug("Stopping the GlusterD Rest server")
	schan := r.srv.StopChan()
	r.srv.Stop(10 * time.Second)
	<-schan
	log.Info("Stopped the GlusterD Rest Server")

	return
}
