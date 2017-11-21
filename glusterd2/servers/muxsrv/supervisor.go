package muxsrv

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc"

	log "github.com/sirupsen/logrus"
	"github.com/thejerf/suture"
)

// New returns a new supervisor managing the mux listener and the multiplexed services
func New() *suture.Supervisor {
	logger := func(msg string) {
		log.WithField("supervisor", "gd2-muxserver").Println(msg)
	}
	s := suture.New("gd2-muxserver", suture.Spec{Log: logger})

	m := newMuxSrv()

	s.Add(rest.NewMuxed(m.m))
	s.Add(sunrpc.NewMuxed(m.m))
	s.Add(m)

	return s
}
