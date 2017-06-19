package elasticetcd

import (
	"os"
	"path"

	"github.com/Sirupsen/logrus"
)

type nilWriteCloser struct{}

func (n *nilWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (n *nilWriteCloser) Close() error {
	return nil
}

func (ee *ElasticEtcd) initLogging() {
	ee.log = logrus.New()
	ee.log.Level = logrus.DebugLevel
	ee.log.Out = new(nilWriteCloser)
	ee.logFile = new(nilWriteCloser)

	if !ee.conf.DisableLogging {
		f, err := os.OpenFile(path.Join(ee.conf.Dir, "elastic.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			ee.log.Out = os.Stdout
			return
		}
		ee.log.Out = f
		ee.logFile = f
	}

	ee.log.Info("beginning etcd logging logging")
}
