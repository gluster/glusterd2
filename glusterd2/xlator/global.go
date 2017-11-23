package xlator

import (
	"fmt"
	"runtime/debug"

	"github.com/gluster/glusterd2/glusterd2/xlator/options"

	log "github.com/sirupsen/logrus"
)

var (
	// xlMap is a map of all available xlators, indexed by xlator-id
	xlMap map[string]*Xlator
	// options is a map of all available options indexed by
	// <xlator-id>.<option-key> for all keys of an option
	optMap map[string]*options.Option
)

// Load load all available xlators and intializes the xlators and options maps
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
	xlMap = xls

	loadOptions()

	return
}

// loadOptions loads all available options into the options.Options map,
// indexed as <xlator-id>.<option-key> for all available option keys
func loadOptions() {
	optMap = make(map[string]*options.Option)
	for _, xl := range xlMap {
		for _, opt := range xl.Options {
			for _, k := range opt.Key {
				k := xl.ID + "." + k
				optMap[k] = opt
			}
		}
	}
}
