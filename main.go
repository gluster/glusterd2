package main

import (
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
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
	logdir, _ := flag.CommandLine.GetString("logdir")
	logFileName, _ := flag.CommandLine.GetString("logfile")

	if logFileName == "-" {
		initLog(logLevel, os.Stderr)
	} else {
		logFilePath := path.Join(logdir, logFileName)
		logFile, logFileErr := openLogFile(logFilePath)
		if logFileErr != nil {
			initLog(logLevel, os.Stderr)
			log.WithError(logFileErr).Fatalf("Failed to open log file %s", logFilePath)
		}
		initLog(logLevel, logFile)
	}

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

	// Start embedded etcd server
	etcdConfig, err := etcdmgmt.GetEtcdConfig(true)
	if err != nil {
		log.WithError(err).Fatal("Could not fetch config options for etcd")
	}
	err = etcdmgmt.StartEmbeddedEtcd(etcdConfig)
	if err != nil {
		log.WithError(err).Fatal("Could not start embedded etcd server")
	}

	// Initialize op version and etcd store
	gdctx.Init()
	if !gdctx.Restart {
		peer.AddSelfDetails()
	}

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
			// Stop embedded etcd server, but don't wipe local etcd data
			etcdmgmt.DestroyEmbeddedEtcd(false)
			super.Stop()
			log.Info("Stopped GlusterD")
			return
		case syscall.SIGHUP:
			// Logrotate case, when Log rotated, Reopen the log file and
			// re-initiate the logger instance.
			log.Info("Received SIGHUP, Reloading log file")
			reloadLog(logdir, logFileName, logLevel)
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
