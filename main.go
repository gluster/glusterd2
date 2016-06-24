package main

import (
	"os"
	"os/signal"

	"github.com/gluster/glusterd2/commands"
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rpc/server"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	config "github.com/spf13/viper"
)

func main() {
	log.WithField("pid", os.Getpid()).Info("GlusterD starting")

	// Parse flags and set up logging before continuing
	parseFlags()
	initLog(logLevel, os.Stderr)

	// Read in config
	initConfig()

	utils.InitDir(config.GetString("localstatedir"))
	context.MyUUID = context.InitMyUUID()

	// Starting etcd daemon upon starting of GlusterD
	etcdCtx, err := etcdmgmt.ETCDStartInit()
	if err != nil {
		log.WithField("Error", err).Fatal("Could not able to start etcd")
	}

	context.Init()
	context.EtcdProcessCtx = etcdCtx

	for _, c := range commands.Commands {
		context.Rest.SetRoutes(c.Routes())
	}

	// Store self information in the store if GlusterD is coming up for
	// first time
	if context.Restart == false {
		peer.AddSelfDetails()
	}

	// Start listening for incoming RPC requests
	err = server.StartListener()
	if err != nil {
		log.Fatal("Could not register RPC listener. Aborting")
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

	// Start GlusterD REST server
	err = context.Rest.Listen()
	if err != nil {
		log.Fatal("Could not start GlusterD Rest Server. Aborting.")
	}

}
