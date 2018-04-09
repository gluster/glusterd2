package events

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/logging"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

const eventLogFileName = "events.log"

var el *eventLogger

type eventLogger struct {
	hID HandlerID
	wc  io.WriteCloser
}

func (l *eventLogger) Handle(e *api.Event) {

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(e); err != nil {
		return
	}

	l.wc.Write(b.Bytes())
}

func (l *eventLogger) Events() []string {
	return nil
}

func startEventLogger() {
	filepath := path.Join(config.GetString(logging.DirFlag), eventLogFileName)
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.WithError(err).WithField("file", filepath).Error("failed to open events log file")
		return
	}

	l := new(eventLogger)
	l.wc = file
	l.hID = Register(l)

	el = l
}

func stopEventLogger() {
	if el == nil {
		return
	}
	Unregister(el.hID)
	el.wc.Close()
	el = nil
}
