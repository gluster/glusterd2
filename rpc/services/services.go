package services

import (
	"net/rpc"
)

// Contains all the list of the rpc services
type service interface{}

var services = []service{
	new(PeerService),
}

func RegisterServices(server *rpc.Server) {
	for _, s := range services {
		//TODO : the service type is as of now int, need to find out a
		// way how to get the type of an object
		server.Register(s)
	}

}
