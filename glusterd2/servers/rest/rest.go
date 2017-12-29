// Package rest implements the REST server for GlusterD
package rest

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"expvar"
	"net"
	"net/http"
	"time"

	"github.com/gluster/glusterd2/glusterd2/middleware"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	config "github.com/spf13/viper"
)

// GDRest is the GlusterD Rest server
type GDRest struct {
	Routes   *mux.Router
	listener net.Listener
	server   *http.Server
	stopCh   chan struct{}
}

func tlsListener(l net.Listener, certfile, keyfile string) (net.Listener, error) {

	certificate, err := tls.LoadX509KeyPair(certfile, keyfile)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		MinVersion:   tls.VersionTLS12, // force TLS 1.2
		Certificates: []tls.Certificate{certificate},
		Rand:         rand.Reader,
	}

	return tls.NewListener(l, config), nil
}

// NewMuxed returns a GDRest object which listens on a CMux multiplexed connection
func NewMuxed(m cmux.CMux) *GDRest {

	rest := &GDRest{
		Routes: mux.NewRouter(),
		server: &http.Server{},
		stopCh: make(chan struct{}),
	}

	certfile := config.GetString("cert-file")
	keyfile := config.GetString("key-file")

	if certfile != "" && keyfile != "" {
		if l, err := tlsListener(m.Match(cmux.TLS()), certfile, keyfile); err != nil {
			// TODO: Don't use Fatal(), bubble up error till main()
			// NOTE: Methods of suture.Service interface do not return error
			log.WithFields(log.Fields{
				"cert-file": certfile,
				"key-file":  keyfile,
			}).WithError(err).Fatal("Failed to create SSL/TLS listener")
		} else {
			rest.listener = l
		}
	} else {
		rest.listener = m.Match(cmux.HTTP1Fast())
	}

	rest.registerRoutes()

	// Expose /statedump endpoint (uses expvar) if enabled
	if ok := config.GetBool("statedump"); ok {
		rest.Routes.Handle("/statedump", expvar.Handler())
	}

	// Chain of ordered middlewares.
	rest.server.Handler = alice.New(
		middleware.Expvar,
		middleware.Recover,
		middleware.ReqIDGenerator,
		middleware.LogRequest,
		middleware.Auth,
	        middleware.Heketi).Then(rest.Routes)

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
