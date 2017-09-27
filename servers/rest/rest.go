// Package rest implements the REST server for GlusterD
package rest

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gluster/glusterd2/middleware"

	log "github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/soheilhy/cmux"
)

// GDRest is the GlusterD Rest server
type GDRest struct {
	Routes   *mux.Router
	listener net.Listener
	server   *http.Server
	stopCh   chan struct{}
}

// NewMuxed returns a GDRest object which listens on a CMux multiplexed connection
func NewMuxed(m cmux.CMux) *GDRest {

	rest := &GDRest{
		Routes:   mux.NewRouter(),
		listener: m.Match(cmux.HTTP1Fast()),
		server:   &http.Server{},
		stopCh:   make(chan struct{}),
	}

	rest.registerRoutes()
	// Chain of ordered middlewares.
	rest.server.Handler = alice.New(
		middleware.Recover,
		middleware.ReqIDGenerator,
		middleware.LogRequest).Then(rest.Routes)

	return rest
}

// Serve begins serving client HTTP requests served by REST server
func (r *GDRest) Serve() {
	log.WithField("ip:port", r.listener.Addr().String()).Info("Started GlusterD ReST server")
	if err := r.server.Serve(r.listener); err != nil && err != cmux.ErrListenerClosed {
		if err == http.ErrServerClosed {
			// when Shutdown() is called, Serve() immediately returns
			// ErrServerClosed. Give Shutdown() a chance to finish.
			<-r.stopCh
		} else {
			log.WithError(err).Error("glusterd ReST server failed")
		}
	}
}

// Stop intends to stop the GlusterD Rest server gracefully. But this won't
// work because the Stop() call chain is managed by supervisor and the cmux
// listener gets closed first.
func (r *GDRest) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	log.Debug("stopping glusterd ReST server gracefully")
	if err := r.server.Shutdown(ctx); err != nil && err != cmux.ErrListenerClosed {
		log.WithError(err).Error("failed to gracefully stop glusterd ReST server")
		if err == context.DeadlineExceeded {
			r.server.Close() // forcefully close connections
		}
	}
	log.Info("stopped glusterd ReST server")

	r.stopCh <- struct{}{}
}
