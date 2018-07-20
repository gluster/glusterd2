// Package rest implements the REST server for GlusterD
package rest

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/gluster/glusterd2/glusterd2/middleware"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/tlsmatcher"

	"github.com/cockroachdb/cmux"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	"go.opencensus.io/plugin/ochttp"
)

const (
	httpReadTimeout  = 10
	httpWriteTimeout = 30
	maxHeaderBytes   = 1 << 13 // 8KB
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
		server: &http.Server{
			ReadTimeout:    httpReadTimeout * time.Second,
			WriteTimeout:   httpWriteTimeout * time.Second,
			MaxHeaderBytes: maxHeaderBytes,
		},
		stopCh: make(chan struct{}),
	}

	certfile := config.GetString("cert-file")
	keyfile := config.GetString("key-file")

	if certfile != "" && keyfile != "" {
		if l, err := tlsListener(m.Match(tlsmatcher.TLS12), certfile, keyfile); err != nil {
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

	// Set Handler to opencensus HTTP handler to enable tracing
	// Set chain of ordered middlewares
	rest.server.Handler = &ochttp.Handler{
		Handler: alice.New(
			middleware.Recover,
			middleware.Expvar,
			middleware.ReqIDGenerator,
			middleware.LogRequest,
			middleware.Auth,
		).Then(rest.Routes),
	}

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

func (r *GDRest) listEndpointsHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var resp api.ListEndpointsResp
		ctx := req.Context()
		for _, r := range AllRoutes {
			resp = append(resp, api.Endpoint{
				Name:         r.Name,
				Method:       r.Method,
				Path:         r.Pattern,
				RequestType:  r.RequestType,
				ResponseType: r.ResponseType,
			})
		}
		restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
	})
}

//Ping URL for glusterd2
func (r *GDRest) Ping() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
	})
}
