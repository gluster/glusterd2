package main

import (
	"github.com/gluster/glusterd2/commands"
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/rpc"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
)

func main() {
	log.Info("GlusterD starting")

	context.Init()

	for _, c := range commands.Commands {
		context.Rest.SetRoutes(c.Routes())
	}
	err := rpc.RegisterServer()
	if err != nil {
		log.Fatal("Could not register as RPC Server. Aborting")
	} else {
		log.Debug("Registered as RPC Server")
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh)
	go func() {
		for s := range sigCh {
			log.WithField("signal", s).Debug("Signal recieved")
			switch s {
			case os.Interrupt:
				log.WithField("signal", s).Info("Recieved SIGTERM. Stopping GlusterD.")
				context.Rest.Stop()
				log.Info("Termintaing GlusterD.")
				os.Exit(0)

			default:
				continue
			}
		}
	}()

	err = context.Rest.Listen()
	if err != nil {
		log.Fatal("Could not start GlusterD Rest Server. Aborting.")
	}

}
