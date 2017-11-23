package xlator

import (
	"fmt"
	"runtime/debug"

	"github.com/gluster/glusterd2/glusterd2/xlator/options"

	log "github.com/sirupsen/logrus"
)

// Xlators is map of all available xlators, indexed by xlator-id
// Other packages can directly import this.
var Xlators map[string]*Xlator

// Load initializes the global variable xlator.Xlators and options.Options
func Load() (err error) {

	defer func() {
		if r := recover(); r != nil {
			log.Info(string(debug.Stack()))
			err = fmt.Errorf("recover()ed at xlator.InitOptions(): %s", r)
			log.Error("You probably didn't install glusterfs from source (branch: experimental)")
		}
	}()

	xls, err := loadAllXlators()
	if err != nil {
		return
	}
	Xlators = xls

	// Also prepare the option.Options map
	loadOptions()

	return
}

// loadOptions loads all available options into the options.Options map,
// indexed as <xlator.ID>.<option.Key> for all available option keys
func loadOptions() {
	for _, xl := range Xlators {
		for _, opt := range xl.Options {
			for _, k := range opt.Key {
				k := xl.ID + "." + k
				options.Options[k] = opt
			}
		}
	}
}
