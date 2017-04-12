package main

import (
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/gapi"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/servers"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	mgmt "github.com/purpleidea/mgmt/lib"
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
	logdir, _ := flag.CommandLine.GetString("logdir")
	logFileName, _ := flag.CommandLine.GetString("logfile")

	initLog(logdir, logFileName, logLevel)

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

	// set all the options we want here...
	libmgmt := &mgmt.Main{}
	libmgmt.Program = "glusterd2"
	libmgmt.Version = "testing" // TODO: set on compilation
	libmgmt.TmpPrefix = true    // prod things probably don't want this on
	//prefix := "/tmp/testprefix/"
	//libmgmt.Prefix = &p // enable for easy debugging
	libmgmt.IdealClusterSize = -1
	libmgmt.ConvergedTimeout = -1
	libmgmt.Noop = false // FIXME: careful!

	libmgmt.GAPI = &gapi.Gd3GAPI{ // graph API
		Program: "gd2",
		Version: "testing",
	}

	if err := libmgmt.Init(); err != nil {
		log.WithError(err).Fatal("Init failed")
	}

	// Initialize op version and etcd store
	gdctx.Init()

	// Start all servers (rest, peerrpc, sunrpc) managed by suture supervisor
	super := initGD2Supervisor()
	super.ServeBackground()
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
		case syscall.SIGHUP:
			// Logrotate case, when Log rotated, Reopen the log file and
			// re-initiate the logger instance.
			if strings.ToLower(logFileName) != "stderr" && strings.ToLower(logFileName) != "stdout" && logFileName != "-" {
				log.Info("Received SIGHUP, Reloading log file")
				initLog(logdir, logFileName, logLevel)
			}
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
