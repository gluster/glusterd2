package main

import (
	"os"
	"os/signal"

	"github.com/gluster/glusterd2/commands"
	"github.com/gluster/glusterd2/config"
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rpc/server"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
)

func main() {
	log.Info("GlusterD starting")

	utils.InitDir(config.LocalStateDir)
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
		transaction.SetTxnSteps(c.Txns())
	}

	// Store self information in the store if GlusterD is coming up for
	// first time
	if context.Restart == false {
		peer.AddSelfDetails()
	}

	err = server.StartListener()
	if err != nil {
		log.Fatal("Could not register the listener. Aborting")
	} else {
		log.Debug("Registered RPC listener")
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
