package log

import (
	"github.com/gluster/glusterd2/pkg/logging"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

func init() {
	var (
		logLvl      = config.GetString("loglevel")
		logdir      = config.GetString("logdir")
		logFileName = config.GetString("logfile")
	)

	if err := logging.Init(logdir, logFileName, logLvl, true); err != nil {
		log.WithError(err).Fatal("failed in initialise logging")
	}
}
