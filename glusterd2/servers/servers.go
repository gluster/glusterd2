// Package servers implements a github.com/thejerf/suture.Supervisor managing all GD2 rpc and rest servers.
package servers

import (
	"github.com/gluster/glusterd2/glusterd2/servers/eventlistener"
	"github.com/gluster/glusterd2/glusterd2/servers/muxsrv"
	"github.com/gluster/glusterd2/glusterd2/servers/peerrpc"

	log "github.com/sirupsen/logrus"
	"github.com/thejerf/suture"
)

// New returns a Supervisor managing the GD2 rpc and rest servers
func New() *suture.Supervisor {
	logger := func(msg string) {
		log.WithField("supervisor", "gd2-servers").Println(msg)
	}

	s := suture.New("gd2-servers", suture.Spec{Log: logger})
	s.Add(peerrpc.New())       // grpc
	s.Add(muxsrv.New())        // sunrpc + http
	s.Add(eventlistener.New()) // eventlistener

	return s
}
