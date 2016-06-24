package server

import (
	"net"
	"net/rpc"

	"github.com/gluster/glusterd2/rpc/services"

	log "github.com/Sirupsen/logrus"
	"github.com/kshlm/pbrpc/pbcodec"
	config "github.com/spf13/viper"
)

// StartListener is to register all the services and start listening on them
func StartListener() error {
	server := rpc.NewServer()
	services.RegisterServices(server)

	listenAddr := config.GetString("rpcaddress")

	l, e := net.Listen("tcp", listenAddr)
	if e != nil {
		log.WithField("error", e).Error("net.Listen() error")
		return e
	} else {
		log.WithField("port", listenAddr).Info("Registered RPC Listener")
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
