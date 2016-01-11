package server

import (
	"net"
	"net/rpc"

	"github.com/gluster/glusterd2/rpc/services"

	log "github.com/Sirupsen/logrus"
	"github.com/kshlm/pbrpc/pbcodec"
)

// StartListener is to register all the services and start listening on them
func StartListener() error {
	server := rpc.NewServer()
	services.RegisterServices(server)
	//TODO : port 9876 is hardcoded now, can be made configurable
	l, e := net.Listen("tcp", ":9876")
	if e != nil {
		log.WithField("error", e).Fatal("listener error")
		return e
	} else {
		log.Debug("listening on port 9876")
	}

	go func() {
		for {
			c, e := l.Accept()
			if e != nil {
				log.WithField("error", e).Info("Accept failed")
				continue
			}
			log.WithField("Connection", c.RemoteAddr()).Info("New incoming connection")
			go server.ServeCodec(pbcodec.NewServerCodec(c))
		}
	}()
	return nil
}
