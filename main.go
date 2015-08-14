package main

import (
	"os"
	"os/signal"

	"github.com/kshlm/glusterd2/commands"
	"github.com/kshlm/glusterd2/context"

	log "github.com/Sirupsen/logrus"
)

func main() {
	log.Info("GlusterD starting")

	context.Init()

	for _, c := range commands.Commands {
		context.Rest.SetRoutes(c.Routes())
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

	err := context.Rest.Listen()
	if err != nil {
		log.Fatal("Could not start GlusterD Rest Server. Aborting.")
	}
}
