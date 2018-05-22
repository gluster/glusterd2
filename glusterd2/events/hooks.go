package events

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

const (
	eventVolumeCreated     = "volume.created"
	eventVolumeOptionSet   = "volume.option.set"
	eventVolumeOptionReset = "volume.option.reset"
	eventVolumeDeleted     = "volume.deleted"
	eventBrickAdded        = "brick.added"
	eventBrickRemoved      = "brick.removed"
)

type hooks struct{}

func (h *hooks) Handle(e *api.Event) {
	var cmd string
	switch e.Name {
	case eventVolumeCreated:
		cmd = "create"
	case eventVolumeStarted:
		cmd = "start"
	case eventVolumeStopped:
		cmd = "stop"
	case eventVolumeOptionSet:
		cmd = "set"
	case eventVolumeOptionReset:
		cmd = "reset"
	case eventVolumeDeleted:
		cmd = "delete"
	case eventBrickAdded:
		cmd = "add-brick"
	case eventBrickRemoved:
		cmd = "remove-brick"
	default:
		return
	}

	hooksPath := fmt.Sprintf("%s/%s/post", config.Get("hooksdir"), cmd)
	hooks := []string{}

	// Read the hooks directory
	hookFiles, err := ioutil.ReadDir(hooksPath)
	if err != nil {
		log.WithError(err).WithField("hooks-dir", hooksPath).Warn("Failed to get list of hook scripts")
		return
	}

	// Collect list of hooks to be executed based on prefix "S"
	// and not symbolic link
	for _, f := range hookFiles {
		if strings.HasPrefix(f.Name(), "S") && f.Mode().IsRegular() {
			hooks = append(hooks, hooksPath+"/"+f.Name())
		}
	}

	// Sort the hooks scripts list
	sort.Slice(hooks, func(i, j int) bool { return hooks[i] < hooks[j] })

	// Execute one by one and record the failures or success
	for _, hook := range hooks {
		if err := utils.ExecuteCommandRun(hook, e.Data["volume.name"]); err != nil {
			log.WithError(err).WithField("command", hook).Warn("Failed to execute hook script")
		} else {
			log.WithField("command", hook).Debug("Hook script succeeded")
		}
	}
}

func (h *hooks) Events() []string {
	return []string{
		eventVolumeCreated,
		eventVolumeStarted,
		eventVolumeStopped,
		eventVolumeDeleted,
		eventVolumeOptionSet,
		eventVolumeOptionReset,
		eventBrickAdded,
		eventBrickRemoved,
	}
}

func registerHooksHandler() {
	h := new(hooks)
	Register(h)
}
