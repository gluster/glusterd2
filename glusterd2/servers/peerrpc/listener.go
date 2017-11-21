// Package peerrpc implements gRPC based GD2-GD2 rpc server
package peerrpc

import (
	"net"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	"google.golang.org/grpc"
)

// Server is the gRPC server
// It provides an implementation of github.com/thejerf/suture.Service interface
type Server struct {
	server *grpc.Server
}

// New returns a new peerrpc.Server with registered gRPC services
func New() *Server {
	s := &Server{
		grpc.NewServer(),
	}
	registerServices(s.server)

	return s
}

// Serve starts a gRPC server
// TODO: This should be able to listen on multiple listeners
func (s *Server) Serve() {
	listenAddr := config.GetString("peeraddress")

	l, e := net.Listen("tcp", listenAddr)
	if e != nil {
		log.WithField("error", e).Error("net.Listen() error")
		return
	}
	log.WithField("ip:port", listenAddr).Info("Registered RPC Listener")

	for svc, si := range s.server.GetServiceInfo() {
		for _, m := range si.Methods {
			log.WithFields(log.Fields{
				"service": svc,
				"method":  m,
			}).Debug("registered gRPC method")
		}
	}

	s.server.Serve(l)
	return
}

// Stop stops the server
func (s *Server) Stop() {
	s.server.GracefulStop()
	return
}
