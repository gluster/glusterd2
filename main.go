package main

import (
	"os"
	"os/signal"

	"github.com/kshlm/glusterd2/commands"
	"github.com/kshlm/glusterd2/config"
	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"

	"github.com/Sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Out = os.Stderr
	logger.Level = logrus.DebugLevel

	logger.Info("GlusterD starting")

	//context := context.New()
	context.Ctx.Log = logger

	context.Ctx.Config = config.New()
	context.Ctx.Config.RestAddress = "localhost:24007"

	context.Ctx.Rest = rest.New(context.Ctx.Config, context.Ctx.Log)

	for _, c := range commands.Commands {
		c.SetRoutes(context.Ctx.Rest.Routes)
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh)
	go func() {
		for s := range sigCh {
			context.Ctx.Log.WithField("signal", s).Debug("Signal recieved")
			switch s {
			case os.Interrupt:
				context.Ctx.Log.WithField("signal", s).Info("Recieved SIGTERM. Stopping GlusterD.")
				context.Ctx.Rest.Stop()
				context.Ctx.Log.Info("Termintaing GlusterD.")
				os.Exit(0)

			default:
				continue
			}
		}
	}()

	err := context.Ctx.Rest.Listen()
	if err != nil {
		logger.Fatal("Could not start GlusterD Rest Server. Aborting.")
	}
}
