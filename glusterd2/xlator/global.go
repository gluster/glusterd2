package xlator

import (
	"fmt"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

// Xlators is map of all available xlators, indexed by xlator-id
// Other packages can directly import this.
var Xlators map[string]*Xlator

// LoadXlators initializes the global variable xlator.AllOptions
func LoadXlators() (err error) {

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
	return
}
