package main

import (
	"os"
	"os/signal"
	"path"
	"strings"
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

	if err := gdctx.SetHostnameAndIP(); err != nil {
		log.WithError(err).Fatal("Failed to get and set hostname or IP")
	}

	// Parse command-line arguments
	parseFlags()

	if showvers, _ := flag.CommandLine.GetBool("version"); showvers {
		dumpVersionInfo()
		return
	}

	logLevel, _ := flag.CommandLine.GetString("loglevel")
	logdir, _ := flag.CommandLine.GetString("logdir")
	logFileName, _ := flag.CommandLine.GetString("logfile")

	if err := initLog(logdir, logFileName, logLevel); err != nil {
		log.WithError(err).Fatal("Failed to initialize logging")
	}

	log.WithField("pid", os.Getpid()).Info("Starting GlusterD")

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

	if err := gdctx.SetUUID(); err != nil {
		log.WithError(err).Fatal("Failed to initialize UUID")
	}

	// Start embedded etcd server
	etcdConfig, err := etcdmgmt.GetEtcdConfig(true)
	if err != nil {
		log.WithError(err).Fatal("Could not fetch config options for etcd")
	}
	if err := etcdmgmt.StartEmbeddedEtcd(etcdConfig); err != nil {
		log.WithError(err).Fatal("Could not start embedded etcd server")
	}

	// Initialize etcd store (etcd client connection)
	if err := gdctx.InitStore(); err != nil {
		log.WithError(err).Fatal("Failed to initialize store (etcd client)")
	}

	if !gdctx.Restart {
		err = peer.AddSelfDetails()
		if err != nil {
			log.WithError(err).Fatal("Could not add self details into etcd")
		}
	}

	// Start all servers (rest, peerrpc, sunrpc) managed by suture supervisor
	super := initGD2Supervisor()
	super.ServeBackground()
	super.Add(servers.New())
	addMgmtService(super)

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
			if strings.ToLower(logFileName) != "stderr" && strings.ToLower(logFileName) != "stdout" && logFileName != "-" {
				log.Info("Received SIGHUP, Reloading log file")
				if err := initLog(logdir, logFileName, logLevel); err != nil {
					log.WithError(err).Fatal("Could not re-initialize logging")
				}
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

func createDirectories() error {
	dirs := []string{config.GetString("localstatedir"),
		config.GetString("rundir"), config.GetString("logdir"),
		path.Join(config.GetString("rundir"), "gluster"),
		path.Join(config.GetString("logdir"), "glusterfs/bricks"),
	}
	for _, dirpath := range dirs {
		if err := utils.InitDir(dirpath); err != nil {
			return err
		}
	}
	return nil
}
