package main

import (
	"os"
	"os/signal"

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

	ctx := context.New()
	ctx.Log = logger

	ctx.Config = config.New()
	ctx.Config.RestAddress = "localhost:24007"

	ctx.Rest = rest.New(ctx.Config, ctx.Log)

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh)
	go func() {
		for s := range sigCh {
			ctx.Log.WithField("signal", s).Debug("Signal recieved")
			switch s {
			case os.Interrupt:
				ctx.Log.WithField("signal", s).Info("Recieved SIGTERM. Stopping GlusterD.")
				ctx.Rest.Stop()
				ctx.Log.Info("Termintaing GlusterD.")
				os.Exit(0)

			default:
				continue
			}
		}
	}()

	err := ctx.Rest.Listen()
	if err != nil {
		logger.Fatal("Could not start GlusterD Rest Server. Aborting.")
	}

}
