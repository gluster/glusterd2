package peerrpc

import (
	"google.golang.org/grpc"
)

// Service is the interface that packages that want to register with the GD2
// RPC server need to implement
type Service interface {
	// RegisterService should register available gRPC Services with the given Server
	// For eg.
	// 	type SomeSvc int
	// 	func (svc *SomeSvc) RegisterService(s *grpc.Server) {
	// 		RegisterSomeSvcServer(s, svc)
	// 	}
	RegisterService(s *grpc.Server)
}

var services []Service

// Register adds the provided Service to the GD2 server services list, which is
// then registered with the grpc.Server.
// Packages need Register during their initialization, to allow all services to
// be registered before StartListener is called.
func Register(svc Service) {
	services = append(services, svc)
}

// registerServices is called by StartListener to get the services list
// registered with the grpc.Server
func registerServices(s *grpc.Server) {
	for _, svc := range services {
		svc.RegisterService(s)
	}
}
