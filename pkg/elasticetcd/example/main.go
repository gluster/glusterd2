// An example application to test and demonstrate elastic-etcd
package main

import (
	"errors"
	"os"
	"os/signal"
	"path"
	"runtime/pprof"

	"github.com/gluster/glusterd2/pkg/elasticetcd"

	"github.com/coreos/etcd/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"golang.org/x/sys/unix"
)

var (
	flags                   elasticetcd.Config
	endpoints, purls, curls []string

	memprof, cpuprof string
)

func init() {
	pflag.StringVar(&flags.Name, "name", "default", "name of this instance")
	pflag.StringVar(&flags.Dir, "datadir", "", "directory to store data")
	pflag.IntVar(&flags.IdealSize, "idealsize", 3, "ideal size of the etcd cluster")
	pflag.StringSliceVar(&endpoints, "endpoints", nil, "endpoints of existing etcd cluster")
	pflag.StringSliceVar(&purls, "purls", nil, "etcd peer urls to listen on")
	pflag.StringSliceVar(&curls, "curls", nil, "etcd client urls to listen on")
}

func parseFlags() {
	pflag.Parse()
}

func getConf() (*elasticetcd.Config, error) {
	var conf elasticetcd.Config

	conf.Name = flags.Name
	conf.IdealSize = flags.IdealSize

	if flags.Dir == "" {
		return nil, errors.New("datadir not given")
	}
	conf.Dir = flags.Dir

	if endpoints != nil {
		urls, err := types.NewURLs(endpoints)
		if err != nil {
			return nil, err
		}
		conf.Endpoints = urls
	}

	if purls != nil {
		urls, err := types.NewURLs(purls)
		if err != nil {
			return nil, err
		}
		conf.PURLs = urls
	}

	if curls != nil {
		urls, err := types.NewURLs(curls)
		if err != nil {
			return nil, err
		}
		conf.CURLs = urls
	}

	return &conf, nil
}

func waitToDie() {
	signals := []os.Signal{unix.SIGTERM, unix.SIGINT}

	sigch := make(chan os.Signal)
	signal.Notify(sigch, signals...)

	sig := <-sigch

	logrus.WithField("signal", sig).Info("got signal")

	return
}

func main() {
	parseFlags()

	cprof, err := os.Create(path.Join(flags.Dir, "cpu.profile"))
	if err == nil {
		pprof.StartCPUProfile(cprof)
		defer pprof.StopCPUProfile()
	}

	conf, err := getConf()
	if err != nil {
		logrus.WithError(err).Fatal("failed to parse options")
	}

	logrus.WithFields(logrus.Fields{
		"name":      conf.Name,
		"datadir":   conf.Dir,
		"endpoints": conf.Endpoints,
		"purls":     conf.PURLs,
		"curls":     conf.CURLs,
		"idealsize": conf.IdealSize,
	}).Info("running elastic etcd with configured options")

	elastic, err := elasticetcd.New(conf)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create new elastic instance")
	}
	logrus.WithField("pid", os.Getpid()).Info("started elastic-etcd")

	waitToDie()

	writeProf()

	elastic.Stop()
}

func writeProf() {
	profiles := []string{"goroutine", "heap", "threadcreate", "block", "mutex"}

	f, err := os.Create(path.Join(flags.Dir, "mem.profile"))
	if err == nil {
		pprof.WriteHeapProfile(f)
		f.Sync()
		f.Close()
	}

	for _, p := range profiles {
		f, err := os.Create(path.Join(flags.Dir, p+".profile"))
		if err == nil {
			pprof.Lookup(p).WriteTo(f, 2)
			f.Sync()
			f.Close()
		}
	}
}
