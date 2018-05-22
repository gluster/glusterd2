package events

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"

	log "github.com/sirupsen/logrus"
)

const (
	eventVolumeStarted = "volume.started"
	eventVolumeStopped = "volume.stopped"
)

const dbusScript = "/usr/libexec/ganesha/dbus-send.sh"

type ganesha struct{}

func (g *ganesha) Handle(e *api.Event) {
	var option string
	if e.Name == eventVolumeStarted {
		option = "on"
	} else if e.Name == eventVolumeStopped {
		option = "off"
	}

	if _, err := os.Stat(dbusScript); os.IsNotExist(err) {
		return
	}
	// TODO: Check if ganesha is running
	dbuscmdStr := fmt.Sprintf("%s /etc/ganesha/ %s %s", dbusScript, option, e.Data["volume.name"])
	if err := utils.ExecuteCommandRun("/bin/sh", "-c", dbuscmdStr); err != nil {
		log.WithError(err).WithField("command", dbuscmdStr).Warn("Failed to execute command")
	} else {
		log.WithField("command", dbuscmdStr).Debug("Command succeeded")
	}
}

func (g *ganesha) Events() []string {
	return []string{eventVolumeStarted, eventVolumeStopped}
}

func registerGaneshaHandler() {
	g := new(ganesha)
	Register(g)
}
