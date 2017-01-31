// Package muxsrv implements a multiplexed TCP server, which multiplexes the GD2 rest, grpc and sunrpc servers
package muxsrv

import (
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/soheilhy/cmux"
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
	mux.l = l
	mux.m = cmux.New(l)

	return mux
}

// Serve starts the handlers and the multiplexed listener
func (m *muxSrv) Serve() {
	log.Info("started muxsrv listener")
	if err := m.m.Serve(); err != nil {
		//TODO: Correctly handle valid errors. We could also be having errors when stopping
		log.WithError(err).Error("mux listener failed")
	}
}

// Stop stops the multiplexed listener and the handlers
func (m *muxSrv) Stop() {
	log.Debug("stopping muxsrv listener")
	if err := m.l.Close(); err != nil {
		log.WithError(err).Error("failed to stop muxsrv listener")
	} else {
		log.Info("stopped muxsrv listener")
	}
}
