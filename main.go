package main

import (
	"os"
	"os/signal"
	"path"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/mgmt"
	"github.com/gluster/glusterd2/servers"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
	"github.com/thejerf/suture"
)

func main() {

	gdctx.SetHostnameAndIP()

	// Parse command-line arguments
	parseFlags()

	if showvers, _ := flag.CommandLine.GetBool("version"); showvers {
		dumpVersionInfo()
		return
	}

	logLevel, _ := flag.CommandLine.GetString("loglevel")
	initLog(logLevel, os.Stderr)

	log.WithField("pid", os.Getpid()).Info("Starting GlusterD")

	// Read config file
	confFile, _ := flag.CommandLine.GetString("config")
	initConfig(confFile)

	workdir := config.GetString("workdir")
	if err := os.Chdir(workdir); err != nil {
		log.WithError(err).Fatalf("Failed to change working directory to %s", workdir)
	}

	// Create directories inside workdir - run dir, logdir etc
	createDirectories()

	// Generate UUID if it doesn't exist
	gdctx.MyUUID = gdctx.InitMyUUID()

	// Start the main supervisor
	super := initGD2Supervisor()
	super.ServeBackground()

	// Initialize op version and etcd store
	gdctx.Init()

	// Start mgmt and the embedded etcd
	m := mgmt.New()
	if m == nil {
		log.Fatal("could not create mgmt service")
	}
	super.Add(m)

	// TODO: Fix once we correctly connect to the store
	//if !gdctx.Restart {
	//peer.AddSelfDetails()
	//}

	// Start all servers (rest, peerrpc, sunrpc) managed by suture supervisor
	super.Add(servers.New())

	// Use the main goroutine as signal handling loop
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh)
	for s := range sigCh {
		log.WithField("signal", s).Debug("Signal received")
		switch s {
		case os.Interrupt:
			log.Info("Received SIGTERM. Stopping GlusterD")
			super.Stop()
			log.Info("Stopped GlusterD")
			return
		default:
			continue
		}
	}
}

func initGD2Supervisor() *suture.Supervisor {
	superlogger := func(msg string) {
		log.WithField("supervisor", "gd2-main").Println(msg)
	}
	return suture.New("gd2-main", suture.Spec{Log: superlogger})
}

func createDirectories() {
	utils.InitDir(config.GetString("localstatedir"))
	utils.InitDir(config.GetString("rundir"))
	utils.InitDir(config.GetString("logdir"))
	utils.InitDir(path.Join(config.GetString("rundir"), "gluster"))
	utils.InitDir(path.Join(config.GetString("logdir"), "glusterfs/bricks"))
}
