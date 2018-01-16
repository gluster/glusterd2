package events

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

const (
	eventVolumeStarted = "volume.started"
	eventVolumeStopped = "volume.stopped"
)

type ganesha struct{}

func (g *ganesha) Handle(e *Event) {
	var option string
	if e.Name == eventVolumeStarted {
		option = "on"
	} else if e.Name == eventVolumeStopped {
		option = "off"
	}
	dbuscmdStr := fmt.Sprintf("/usr/libexec/ganesha/dbus-send.sh /etc/ganesha/ %s %s", option, e.Data["volume.name"])
	ganeshaCmd := exec.Command("/bin/sh", "-c", dbuscmdStr)
	err := ganeshaCmd.Run()
	if err != nil {
		log.WithError(err).WithField("command", dbuscmdStr).Error("Failed to execute command")
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
