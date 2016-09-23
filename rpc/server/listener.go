package server

import (
	"net"

	log "github.com/Sirupsen/logrus"
	config "github.com/spf13/viper"
	"google.golang.org/grpc"
)

// StartListener is to register all the services and start listening on them
func StartListener() error {
	server := grpc.NewServer()

	listenAddr := config.GetString("rpcaddress")

	l, e := net.Listen("tcp", listenAddr)
	if e != nil {
		log.WithField("error", e).Error("net.Listen() error")
		return e
	} else {
		log.WithField("ip:port", listenAddr).Info("Registered RPC Listener")
	}

	server.Serve(l)
	return nil
}
