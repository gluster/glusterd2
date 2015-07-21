// Package rest implements the REST server for GlusterD
package rest

import (
	"net/http"
	"time"

	"github.com/kshlm/glusterd2/config"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"gopkg.in/tylerb/graceful.v1"
)

type GDRest struct {
	Routes *mux.Router

	srv *graceful.Server

	logger *logrus.Logger
}

// function New returns a GlusterDRest object which listens on the give address
func New(cfg *config.GDConfig, logger *logrus.Logger) *GDRest {
	rest := &GDRest{}

	rest.Routes = mux.NewRouter()

	n := negroni.New()
	n.Use(&negronilogrus.Middleware{Logger: logger, Name: "glusterd-rest"})
	n.UseHandler(rest.Routes)

	rest.srv = &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:    cfg.RestAddress,
			Handler: n,
		},
	}

	rest.logger = logger

	return rest
}

func (r *GDRest) Listen() error {
	r.logger.Debug("Beginning the GlusterD Rest server")
	err := r.srv.ListenAndServe()
	if err != nil {
		r.logger.WithField("error", err).Error("Failed to start the Rest Server")
		return err
	}
	return nil
}

func (r *GDRest) Stop() {
	r.logger.Debug("Stopping the GlusterD Rest server")
	schan := r.srv.StopChan()
	r.srv.Stop(10 * time.Second)
	<-schan
	r.logger.Info("Stopped the GlusterD Rest Server")

	return
}
