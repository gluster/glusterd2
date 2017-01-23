package main

import (
	"net"
	"os"
	"os/signal"
	"path"

	"github.com/gluster/glusterd2/commands"
	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rpc/server"
	"github.com/gluster/glusterd2/rpc/sunrpcserver"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/prashanthpai/sunrpc"
	"github.com/soheilhy/cmux"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

func main() {

	// Set IP and hostname once.
	gdctx.SetHostnameAndIP()

	// Parse flags and handle version and logging before continuing
	parseFlags()

	showvers, _ := flag.CommandLine.GetBool("version")
	if showvers {
		dumpVersionInfo()
		return
	}

	logLevel, _ := flag.CommandLine.GetString("loglevel")
	initLog(logLevel, os.Stderr)

	log.WithField("pid", os.Getpid()).Info("GlusterD starting")

	// Read in config
	confFile, _ := flag.CommandLine.GetString("config")
	initConfig(confFile)

	// Change to working directory before continuing
	if e := os.Chdir(config.GetString("workdir")); e != nil {
		log.WithError(e).Fatalf("failed to change working directory")
	}

	// TODO: This really should go into its own function.
	utils.InitDir(config.GetString("localstatedir"))
	utils.InitDir(config.GetString("rundir"))
	utils.InitDir(config.GetString("logdir"))
	utils.InitDir(path.Join(config.GetString("rundir"), "gluster"))
	utils.InitDir(path.Join(config.GetString("logdir"), "glusterfs/bricks"))

	gdctx.MyUUID = gdctx.InitMyUUID()

	// Start embedded etcd server
	etcdConfig, err := etcdmgmt.GetEtcdConfig(true)
	if err != nil {
		log.WithField("Error", err).Fatal("Could not fetch config options for etcd.")
	}
	err = etcdmgmt.StartEmbeddedEtcd(etcdConfig)
	if err != nil {
		log.WithField("Error", err).Fatal("Could not start embedded etcd server.")
	}

	gdctx.Init()

	for _, c := range commands.Commands {
		gdctx.Rest.SetRoutes(c.Routes())
		c.RegisterStepFuncs()
	}

	// Store self information in the store if GlusterD is coming up for
	// first time
	if !gdctx.Restart {
		peer.AddSelfDetails()
	}

	// Start listening for incoming RPC requests from other peers
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
				gdctx.Rest.Stop()
				etcdmgmt.DestroyEmbeddedEtcd()
				server.StopServer()
				log.Info("Termintaing GlusterD.")
				os.Exit(0)

			default:
				continue
			}
		}
	}()

	// Main listener for all client requests
	l, err := net.Listen("tcp", config.GetString("clientaddress"))
	if err != nil {
		log.Fatal(err)
	}

	// Create a mux
	m := cmux.New(l)

	// Match connections
	httpL := m.Match(cmux.HTTP1Fast())
	sunrpcL := m.Match(sunrpc.CmuxMatcher())

	// Start REST server and listen to HTTP requests from clients
	err = gdctx.Rest.Serve(httpL)
	if err != nil {
		log.Fatal("Could not start GlusterD Rest Server. Aborting.")
	}

	// Start Sun RPC server and listen to requests from glusterfs clients
	err = sunrpcserver.Start(sunrpcL)
	if err != nil {
		log.Fatal("Could not start Sun RPC server. Aborting.")
	}

	// Start serving client requests. This will start multiplexing the
	// listener and shall block.
	m.Serve()
}
