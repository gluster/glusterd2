// Package peerrpc implements gRPC based GD2-GD2 rpc server
package peerrpc

import (
	"net"

	log "github.com/Sirupsen/logrus"
	config "github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func init() {
	// The GD2 logger for gRPC
	grpclog.SetLogger(log.StandardLogger().WithField("module", "gRPC"))
}

// PeerRPCServer is the gRPC server
// It provides an implementation of github.com/thejerf/suture.Service interface
type PeerRPCServer struct {
	server *grpc.Server
}

// New returns a new PeerRPCServer
func New() *PeerRPCServer {
	return new(PeerRPCServer)
}

// Serve registers gRCP services and starts a gRPC listener
// TODO: This should be able to listen on multiple listeners
func (s *PeerRPCServer) Serve() {
	s.server = grpc.NewServer()
	registerServices(s.server)

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
func (s *PeerRPCServer) Stop() {
	s.server.GracefulStop()
	return
}
