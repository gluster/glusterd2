// Package muxsrv implements a multiplexed TCP server, which multiplexes the GD2 rest, grpc and sunrpc servers
package muxsrv

import (
	"net"

	"github.com/gluster/glusterd2/servers/rest"

	log "github.com/Sirupsen/logrus"
	"github.com/soheilhy/cmux"
	config "github.com/spf13/viper"
	"github.com/thejerf/suture"
)

// MuxSrv implements the suture.Sever for the GD2 multiplexed server
type MuxSrv struct {
	*suture.Supervisor
	l net.Listener
	m cmux.CMux
}

// New returns a multiplexed server with the multiplexed listeners already setup
func New() *MuxSrv {
	mux := &MuxSrv{}
	logger := func(msg string) {
		log.WithField("supervisor", "gd2-muxserver").Println(msg)
	}
	mux.Supervisor = suture.New("gd2-muxserver", suture.Spec{Log: logger})

	l, err := net.Listen("tcp", config.GetString("clientaddress"))
	if err != nil {
		log.Fatal(err)
	}
	mux.l = l
	mux.m = cmux.New(l)

	mux.Supervisor.Add(rest.NewMuxed(mux.m))

	return mux
}

// Serve starts the handlers and the multiplexed listener
func (m *MuxSrv) Serve() {
	go m.Supervisor.ServeBackground()
	m.m.Serve()
	return
}

// Stop stops the multiplexed listener and the handlers
func (m *MuxSrv) Stop() {
	m.l.Close()
	m.Supervisor.Stop()
}
