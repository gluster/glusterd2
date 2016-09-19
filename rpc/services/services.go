package services

import (
	"net/rpc"

	"github.com/gluster/glusterd2/transaction"
)

// Contains all the list of the rpc services
type service interface{}

var services = []service{
	new(PeerService),
	new(transaction.TxnSvc),
}

func RegisterServices(server *rpc.Server) {
	for _, s := range services {
		//TODO : the service type is as of now int, need to find out a
		// way how to get the type of an object
		server.Register(s)
	}

}
