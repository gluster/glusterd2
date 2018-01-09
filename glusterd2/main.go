package main

import (
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/gluster/glusterd2/glusterd2/commands/volumes"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/glusterd2/servers"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/logging"
	"github.com/gluster/glusterd2/pkg/utils"
	"github.com/gluster/glusterd2/version"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
	"github.com/thejerf/suture"
	"golang.org/x/sys/unix"
)

func main() {

	if err := gdctx.SetHostnameAndIP(); err != nil {
		log.WithError(err).Fatal("Failed to get and set hostname or IP")
	}

	// Parse command-line arguments
	parseFlags()

	if showvers, _ := flag.CommandLine.GetBool("version"); showvers {
		version.DumpVersionInfo()
		return
	}

	logLevel, _ := flag.CommandLine.GetString("loglevel")
	logdir, _ := flag.CommandLine.GetString("logdir")
	logFileName, _ := flag.CommandLine.GetString("logfile")

	if err := logging.Init(logdir, logFileName, logLevel, true); err != nil {
		log.WithError(err).Fatal("Failed to initialize logging")
	}

	log.WithFields(log.Fields{
		"pid":     os.Getpid(),
		"version": version.GlusterdVersion,
	}).Debug("Starting GlusterD")

	// Read config file
	confFile, _ := flag.CommandLine.GetString("config")
	if err := initConfig(confFile); err != nil {
		log.WithError(err).Fatal("Failed to initialize config")
	}

	workdir := config.GetString("workdir")
	if err := os.Chdir(workdir); err != nil {
		log.WithError(err).Fatalf("Failed to change working directory to %s", workdir)
	}

	// Create directories inside workdir - run dir, logdir etc
	if err := createDirectories(); err != nil {
		log.WithError(err).Fatal("Failed to create or access directories")
	}

	if err := gdctx.InitUUID(); err != nil {
		log.WithError(err).Fatal("Failed to initialize UUID")
	}

	// Load all possible xlator options
	if err := xlator.Load(); err != nil {
		log.WithError(err).Fatal("Failed to load xlator options")
	}

	// Load volgen templates
	if err := volgen.LoadTemplates(); err != nil {
		log.WithError(err).Fatal("Failed to load volgen templates")
	}

	// Initialize etcd store (etcd client connection)
	if err := store.Init(nil); err != nil {
		log.WithError(err).Fatal("Failed to initialize store (etcd client)")
	}

	// Start the events framework after store is up
	if err := events.Start(); err != nil {
		log.WithError(err).Fatal("Failed to start internal events framework")
	}

	if err := peer.AddSelfDetails(); err != nil {
		log.WithError(err).Fatal("Could not add self details into etcd")
	}

	// Load the default group option map into the store
	if err := volumecommands.LoadDefaultGroupOptions(); err != nil {
		log.WithError(err).Fatal("Failed to load the default group options")
	}

	// If REST API Auth is enabled, Generate Auth file with random secret in workdir
	if err := gdctx.GenerateLocalAuthToken(); err != nil {
		log.WithError(err).Fatal("Failed to generate local auth token")
	}

	// Start all servers (rest, peerrpc, sunrpc) managed by suture supervisor
	super := initGD2Supervisor()
	super.ServeBackground()
	super.Add(servers.New())

	// Restart previously running daemons
	daemon.StartAllDaemons()

	// Use the main goroutine as signal handling loop
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh)
	for s := range sigCh {
		log.WithField("signal", s).Debug("Signal received")
		switch s {
		case unix.SIGTERM:
			fallthrough
		case unix.SIGINT:
			log.Info("Received SIGTERM. Stopping GlusterD")
			super.Stop()
			events.Stop()
			store.Close()
			log.Info("Stopped GlusterD")
			return
		case unix.SIGHUP:
			// Logrotate case, when Log rotated, Reopen the log file and
			// re-initiate the logger instance.
			if strings.ToLower(logFileName) != "stderr" && strings.ToLower(logFileName) != "stdout" && logFileName != "-" {
				log.Info("Received SIGHUP, Reloading log file")
				if err := logging.Init(logdir, logFileName, logLevel, true); err != nil {
					log.WithError(err).Fatal("Could not re-initialize logging")
				}
			}
		case unix.SIGUSR1:
			log.Info("Received SIGUSR1. Dumping statedump")
			utils.WriteStatedump()
		default:
			continue
		}
	}
}

func initGD2Supervisor() *suture.Supervisor {
	superlogger := func(msg string) {
		log.WithField("supervisor", "gd2-main").Println(msg)
	}
	return suture.New("gd2-main", suture.Spec{Log: superlogger, Timeout: 5 * time.Second})
}

func createDirectories() error {
	dirs := []string{config.GetString("localstatedir"),
		config.GetString("rundir"), config.GetString("logdir"),
		path.Join(config.GetString("rundir"), "gluster"),
		path.Join(config.GetString("logdir"), "glusterfs/bricks"),
		"/var/run/gluster", // issue #476
	}
	for _, dirpath := range dirs {
		if err := utils.InitDir(dirpath); err != nil {
			return err
		}
	}
	return nil
}
