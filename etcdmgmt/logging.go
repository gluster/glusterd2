package etcdmgmt

import (
	"io"
	"os"
	"path"

	"github.com/coreos/pkg/capnslog"
	config "github.com/spf13/viper"
)

var etcdLogWriter io.WriteCloser

func initEtcdLogging() error {
	etcdLogFile := path.Join(config.GetString("logdir"), config.GetString("etcdlogfile"))
	f, err := os.OpenFile(etcdLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	etcdLogWriter = f
	capnslog.SetFormatter(capnslog.NewPrettyFormatter(etcdLogWriter, false))
	return nil
}
