package server

import (
	"net"

	log "github.com/Sirupsen/logrus"
	config "github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	server *grpc.Server
)

func init() {
	// The GD2 logger for gRPC
	grpclog.SetLogger(log.StandardLogger().WithField("module", "gRPC"))
}

// StartListener is to register all the services and start listening on them
// TODO: This should be able to listen on multiple listeners
func StartListener() error {
	server = grpc.NewServer()
	registerServices(server)

	listenAddr := config.GetString("peeraddress")

	l, e := net.Listen("tcp", listenAddr)
	if e != nil {
		log.WithField("error", e).Error("net.Listen() error")
		return e
	}
	log.WithField("ip:port", listenAddr).Info("Registered RPC Listener")

	for s, si := range server.GetServiceInfo() {
		for _, m := range si.Methods {
			log.WithFields(log.Fields{
				"service": s,
				"method":  m,
			}).Debug("registered gRPC method")
		}
	}

	go server.Serve(l)
	return nil
}

// StopServer stops the server
func StopServer() {
	server.GracefulStop()
}
