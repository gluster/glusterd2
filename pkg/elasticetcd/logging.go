package elasticetcd

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// TODO: Allow custom logging locations

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
		if err := os.MkdirAll(ee.conf.LogDir, 0755); err != nil {
			// Log to stdout if you can create log file
			ee.log.Out = os.Stdout
			return
		}
		f, err := os.OpenFile(path.Join(ee.conf.LogDir, "elastic.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// Log to stdout if you can create log file
			ee.log.Out = os.Stdout
			return
		}
		ee.log.Out = f
		ee.logFile = f
	}
	return
}
