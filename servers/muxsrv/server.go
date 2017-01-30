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
		log.Fatal(err)
	}
	mux.l = l
	mux.m = cmux.New(l)

	return mux
}

// Serve starts the handlers and the multiplexed listener
func (m *muxSrv) Serve() {
	m.m.Serve()
	return
}

// Stop stops the multiplexed listener and the handlers
func (m *muxSrv) Stop() {
	m.l.Close()
}
