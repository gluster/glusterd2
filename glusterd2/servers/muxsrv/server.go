// Package muxsrv implements a multiplexed TCP server, which multiplexes the GD2 rest, grpc and sunrpc servers
package muxsrv

import (
	"crypto/rand"
	"crypto/tls"
	"net"

	"github.com/gluster/glusterd2/constants"

	"github.com/cockroachdb/cmux"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

// MuxSrv implements the suture.Sever for the GD2 multiplexed server
type muxSrv struct {
	l net.Listener
	m cmux.CMux
}

// newMuxSrv returns a multiplexed server with the multiplexed listeners already setup
func newMuxSrv() *muxSrv {
	mux := &muxSrv{}

	l, err := net.Listen("tcp", config.GetString("clientaddress"))
	if err != nil {
		log.WithError(err).Fatal("failed to create gd2-muxsrv listener")
	}

	if config.GetBool(constants.UseTLS) {
		cert := config.GetString(constants.ClntCertFile)
		key := config.GetString(constants.ClntKeyFile)

		if l, err = tlsListener(l, cert, key); err != nil {
			// TODO: Don't use Fatal(), bubble up error till main()
			// NOTE: Methods of suture.Service interface do not return error
			log.WithFields(log.Fields{
				"cert-file": cert,
				"key-file":  key,
			}).WithError(err).Fatal("failed to create TLS client listener")
		}
	}
	mux.l = l
	mux.m = cmux.New(l)

	return mux
}

// tlsListener returns a TLS listener configured using the given certificate
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

// Serve starts the handlers and the multiplexed listener
func (m *muxSrv) Serve() {
	if err := m.m.Serve(); err != nil && err != cmux.ErrListenerClosed {
		log.WithError(err).Warn("mux listener failed")
	}
}

// Stop stops the multiplexed listener and the handlers
func (m *muxSrv) Stop() {
	if err := m.l.Close(); err != nil && err != cmux.ErrListenerClosed {
		log.WithError(err).Warn("failed to stop muxsrv listener")
	} else {
		log.Info("stopped muxsrv listener")
	}
}
