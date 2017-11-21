package xlator

import (
	"fmt"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

// AllOptions contains all possible xlator options for all xlators
// Other packages can directly import this.
// The keys are of the form <xlator>.<option>
// Example: afr.eager-lock
var AllOptions map[string][]Option

// InitOptions initializes the global variable xlator.AllOptions
func InitOptions() (err error) {

	defer func() {
		if r := recover(); r != nil {
			log.Info(string(debug.Stack()))
			err = fmt.Errorf("recover()ed at xlator.InitOptions(): %s", r)
			log.Error("You probably didn't install glusterfs from source (branch: experimental)")
		}
	}()

	xopts, err := getAllOptions()
	if err != nil {
		return
	}
	AllOptions = xopts
	return
}
