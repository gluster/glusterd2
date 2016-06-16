package transaction

// The registry registers StepFunc's

import (
	"errors"
	"reflect"
	"sync"

	"github.com/gluster/glusterd2/context"

	log "github.com/Sirupsen/logrus"
)

var (
	stepRegistry map[string]StepFunc
	srMutex      sync.Mutex
)

func init() {
	stepRegistry = make(map[string]StepFunc)
}

func registerStepFunc(s *StepFunc, name string) {
	if _, ok := stepRegistry[name]; !ok {
		log.WithField("stepname", name).Warning("step with provided name exists in registry and will be overwritten")
	}

	stepRegistry[name] = s
}

//RegisterStepFunc registers the given StepFunc in the registry
func RegisterStep(s StepFunc, name string) {
	srMutex.Lock()
	defer srMutex.Unlock()

	registerStep(s, name)
}

//GetStepFunc returns named step if found.
func GetStepFunc(name string) (StepFunc, bool) {
	srMutex.Lock()
	defer srMutex.Unlock()

	s, ok := stepRegistry[name]
	return s, ok
}
