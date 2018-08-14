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
	gdutils "github.com/gluster/glusterd2/glusterd2/utils"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/logging"
	"github.com/gluster/glusterd2/pkg/tracing"
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

	// Initalize and parse CLI flags
	initFlags()

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

	// Initialize GD2 config
	if err := initConfig(); err != nil {
		log.WithError(err).Fatal("Failed to initialize config")
	}

	logLevel2 := config.GetString("loglevel")
	logdir2 := config.GetString("logdir")
	logFileName2 := config.GetString("logfile")

	if logLevel != logLevel2 || logdir != logdir2 || logFileName != logFileName2 {
		if err := logging.Init(logdir2, logFileName2, logLevel2, true); err != nil {
			log.WithError(err).Fatal("Failed to re-initialize logging")
		}
	}

	log.WithFields(log.Fields{
		"pid":     os.Getpid(),
		"version": version.GlusterdVersion,
	}).Debug("Starting GlusterD")

	dumpConfigToLog()

	workdir := config.GetString("localstatedir")
	if err := os.Chdir(workdir); err != nil {
		log.WithError(err).Fatalf("Failed to change working directory to %s", workdir)
	}

	// Create directories inside localstatedir - run dir, logdir etc
	if err := createDirectories(); err != nil {
		log.WithError(err).Fatal("Failed to create or access directories")
	}

	// Create pidfile if specified
	if err := createPidFile(); err != nil {
		log.WithError(err).Fatal("Failed to create pid file")
	}

	if err := gdctx.InitUUID(); err != nil {
		log.WithError(err).Fatal("Failed to initialize UUID")
	}

	// Load all possible xlator options
	if err := xlator.Load(); err != nil {
		log.WithError(err).Fatal("Failed to load xlator options")
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

	// If REST API Auth is enabled, Generate Auth file with random secret in localstatedir
	if err := gdctx.GenerateLocalAuthToken(); err != nil {
		log.WithError(err).Fatal("Failed to generate local auth token")
	}

	// Create the Opencensus Jaeger exporter
	if exporter := tracing.InitJaegerExporter(); exporter != nil {
		defer exporter.Flush()
	}

	// Start all servers (rest, peerrpc, sunrpc) managed by suture supervisor
	super := initGD2Supervisor()
	super.ServeBackground()
	super.Add(servers.New())

	// Restart previously running daemons
	daemon.StartAllDaemons()

	// Mount all Local Bricks
	gdutils.MountLocalBricks()

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
			_ = os.Remove(config.GetString("pidfile"))
			log.Info("Stopped GlusterD")
			return
		case unix.SIGHUP:
			// Logrotate case, when Log rotated, Reopen the log file and
			// re-initiate the logger instance.
			if strings.ToLower(logFileName2) != "stderr" && strings.ToLower(logFileName2) != "stdout" && logFileName2 != "-" {
				log.Info("Received SIGHUP, Reloading log file")
				if err := logging.Init(logdir2, logFileName2, logLevel2, true); err != nil {
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
		path.Join(config.GetString("logdir"), "glusterfs/bricks"),
		path.Join(config.GetString("hooksdir"), "create/post"),
		path.Join(config.GetString("hooksdir"), "start/post"),
		path.Join(config.GetString("hooksdir"), "stop/post"),
		path.Join(config.GetString("hooksdir"), "set/post"),
		path.Join(config.GetString("hooksdir"), "reset/post"),
		path.Join(config.GetString("hooksdir"), "delete/post"),
		path.Join(config.GetString("hooksdir"), "add-brick/post"),
		path.Join(config.GetString("hooksdir"), "remove-brick/post"),
		"/var/run/gluster", // issue #476
	}
	for _, dirpath := range dirs {
		if err := utils.InitDir(dirpath); err != nil {
			return err
		}
	}
	return nil
}

func createPidFile() error {
	pidfile := config.GetString("pidfile")

	// Check if pidfile exists and already running
	pid, err := daemon.ReadPidFromFile(pidfile)
	if err == nil {
		// Check if process is running
		_, err := daemon.GetProcess(pid)
		if err == nil {
			return errors.ErrProcessAlreadyRunning
		}
	}

	return daemon.WritePidToFile(os.Getpid(), pidfile)
}
